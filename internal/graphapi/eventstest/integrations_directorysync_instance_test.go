//go:build test

package eventstest_test

import (
	"context"
	"encoding/json"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	openapi "github.com/theopenlane/core/common/openapi"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorygroup"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorymembership"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	intregistry "github.com/theopenlane/core/v2/internal/integrations/registry"
	integrationtypes "github.com/theopenlane/core/v2/internal/integrations/types"
)

// ingestDirectorySnapshot runs one full sync batch (accounts, groups, memberships) through the
// synchronous ingest path, exactly as a reconcile cycle applies a provider snapshot
func ingestDirectorySnapshot(ctx context.Context, t *testing.T, integration *ent.Integration, accounts, groups, memberships []string) operations.IngestResult {
	t.Helper()

	def := directorySyncTestDefinition(integration.DefinitionID)
	reg := intregistry.New()
	th.RequireNoError(t, reg.Register(def))

	toEnvelopes := func(payloads []string) []integrationtypes.MappingEnvelope {
		envelopes := make([]integrationtypes.MappingEnvelope, 0, len(payloads))
		for _, p := range payloads {
			envelopes = append(envelopes, integrationtypes.MappingEnvelope{Payload: json.RawMessage(p)})
		}

		return envelopes
	}

	result, err := operations.ProcessPayloadSets(ctx, operations.IngestContext{
		Registry:    reg,
		DB:          suite.Client.DB,
		Integration: integration,
	}, directorySyncTestOperation, def.Operations[0].Ingest, def.Operations[0].Policy, []integrationtypes.IngestPayloadSet{
		{Schema: entityops.SchemaDirectoryAccount.Name, Envelopes: toEnvelopes(accounts)},
		{Schema: entityops.SchemaDirectoryGroup.Name, Envelopes: toEnvelopes(groups)},
		{Schema: entityops.SchemaDirectoryMembership.Name, Envelopes: toEnvelopes(memberships), SnapshotComplete: true},
	}, operations.IngestOptions{})
	th.RequireNoError(t, err)

	return result
}

func TestDirectorySyncInstanceScopedCorrelation(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	first, err := suite.Client.DB.Integration.Create().
		SetName("Instance Correlation Test").
		SetKind("dirinsttest").
		SetDefinitionID("def_dirinsttest").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-1"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	reinstalled, err := suite.Client.DB.Integration.Create().
		SetName("Instance Correlation Test Reinstalled").
		SetKind("dirinsttest").
		SetDefinitionID("def_dirinsttest").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-1"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	t.Cleanup(func() {
		for _, integrationID := range []string{first.ID, reinstalled.ID} {
			_, err := suite.Client.DB.DirectoryMembership.Delete().Where(directorymembership.IntegrationID(integrationID)).Exec(ctx)
			th.RequireNoError(t, err)
		}

		_, err := suite.Client.DB.DirectoryAccount.Delete().Where(directoryaccount.ExternalIDHasPrefix("dirinst-")).Exec(ctx)
		th.RequireNoError(t, err)
		_, err = suite.Client.DB.DirectoryGroup.Delete().Where(directorygroup.ExternalIDHasPrefix("dirinst-")).Exec(ctx)
		th.RequireNoError(t, err)

		th.RequireNoError(t, suite.Client.DB.Integration.DeleteOneID(first.ID).Exec(ctx))
		th.RequireNoError(t, suite.Client.DB.Integration.DeleteOneID(reinstalled.ID).Exec(ctx))
	})

	accounts := []string{`{"external_id":"dirinst-a-1","canonical_email":"dirinst1@example.com","display_name":"Instance User","directory_name":"Instance Correlation Test","profile":{"id":"dirinst-a-1","rev":1}}`}
	groups := []string{`{"external_id":"dirinst-g-1","display_name":"Instance Group","profile":{"id":"dirinst-g-1","rev":1}}`}
	memberships := []string{`{"directory_account_id":"dirinst-a-1","directory_group_id":"dirinst-g-1"}`}

	seeded := ingestDirectorySnapshot(ctx, t, first, accounts, groups, memberships)
	assert.Check(t, is.Equal(0, seeded.Failed))
	assert.Check(t, is.Equal(3, seeded.Changed), "the first sync must create all three rows")

	created, err := suite.Client.DB.DirectoryAccount.Query().Where(directoryaccount.ExternalID("dirinst-a-1")).Only(ctx)
	th.RequireNoError(t, err)
	assert.Check(t, is.Equal("tenant-1", created.SourceInstanceID), "the backfill sync must stamp the existing row instead of duplicating it")

	resynced := ingestDirectorySnapshot(ctx, t, reinstalled, accounts, groups, memberships)
	assert.Check(t, is.Equal(0, resynced.Failed))
	assert.Check(t, is.Equal(0, resynced.Changed), "a second live installation must not change rows the first one manages")
	assert.Check(t, is.Equal(3, resynced.Skipped), "a second live installation must skip every row the first one manages")

	confirmed, err := suite.Client.DB.DirectoryAccount.Query().Where(directoryaccount.ExternalID("dirinst-a-1")).Only(ctx)
	th.RequireNoError(t, err)
	assert.Check(t, is.Equal("Instance Correlation Test", *confirmed.DirectoryName), "the confirming integration's display name must not overwrite the shared row")

	accountCount, err := suite.Client.DB.DirectoryAccount.Query().Where(directoryaccount.ExternalID("dirinst-a-1")).Count(ctx)
	th.RequireNoError(t, err)
	assert.Check(t, is.Equal(1, accountCount), "a reinstalled integration sharing the instance must not duplicate the account")

	groupCount, err := suite.Client.DB.DirectoryGroup.Query().Where(directorygroup.ExternalID("dirinst-g-1")).Count(ctx)
	th.RequireNoError(t, err)
	assert.Check(t, is.Equal(1, groupCount), "a reinstalled integration sharing the instance must not duplicate the group")

	group, err := suite.Client.DB.DirectoryGroup.Query().Where(directorygroup.ExternalID("dirinst-g-1")).Only(ctx)
	th.RequireNoError(t, err)

	membership, err := suite.Client.DB.DirectoryMembership.Query().
		Where(directorymembership.DirectoryAccountID(created.ID)).
		Where(directorymembership.DirectoryGroupID(group.ID)).
		Only(ctx)
	th.RequireNoError(t, err)
	assert.Check(t, is.Equal(first.ID, membership.IntegrationID), "the shared active membership keeps its creating integration's attribution")
}
