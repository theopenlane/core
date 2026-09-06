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
	"github.com/theopenlane/core/v2/internal/ent/generated/directorysyncrun"
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
	}, directorySyncTestOperation, def.Operations[0].Ingest, []integrationtypes.IngestPayloadSet{
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

		for _, integrationID := range []string{first.ID, reinstalled.ID} {
			_, err := suite.Client.DB.DirectorySyncRun.Delete().Where(directorysyncrun.IntegrationID(integrationID)).Exec(ctx)
			th.RequireNoError(t, err)
		}

		th.RequireNoError(t, suite.Client.DB.Integration.DeleteOneID(first.ID).Exec(ctx))
		th.RequireNoError(t, suite.Client.DB.Integration.DeleteOneID(reinstalled.ID).Exec(ctx))
	})

	accounts := []string{`{"external_id":"dirinst-a-1","canonical_email":"dirinst1@example.com","display_name":"Instance User","profile":{"id":"dirinst-a-1","rev":1}}`}
	groups := []string{`{"external_id":"dirinst-g-1","display_name":"Instance Group","profile":{"id":"dirinst-g-1","rev":1}}`}
	memberships := []string{`{"directory_account_id":"dirinst-a-1","directory_group_id":"dirinst-g-1"}`}

	seeded := ingestDirectorySnapshot(ctx, t, first, accounts, groups, memberships)
	assert.Check(t, is.Equal(0, seeded.Failed))
	assert.Check(t, is.Equal(3, seeded.Changed), "the first sync must create all three rows")

	// the installation resolves its instance identity, and the next sync backfills existing rows
	first, err = suite.Client.DB.Integration.UpdateOneID(first.ID).
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-1"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	backfilled := ingestDirectorySnapshot(ctx, t, first, accounts, groups, memberships)
	assert.Check(t, is.Equal(0, backfilled.Failed))
	assert.Check(t, is.Equal(3, backfilled.Changed), "a sync after the installation resolves its instance must backfill each existing row once")

	created, err := suite.Client.DB.DirectoryAccount.Query().Where(directoryaccount.ExternalID("dirinst-a-1")).Only(ctx)
	th.RequireNoError(t, err)
	assert.Check(t, is.Equal("tenant-1", created.SourceInstanceID), "the backfill sync must stamp the existing row instead of duplicating it")

	resynced := ingestDirectorySnapshot(ctx, t, reinstalled, accounts, groups, memberships)
	assert.Check(t, is.Equal(0, resynced.Failed))
	assert.Check(t, is.Equal(0, resynced.Changed), "a reinstalled integration confirming correlated rows must report no changes")

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
	assert.Check(t, membership.LastConfirmedRunID != nil, "the reinstalled integration's run must confirm the shared membership")
}

func TestDirectorySyncDuplicateRowConvergence(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	one, err := suite.Client.DB.Integration.Create().
		SetName("Dup Convergence A").
		SetKind("dirduptest").
		SetDefinitionID("def_dirduptest").
		Save(ctx)
	th.RequireNoError(t, err)

	other, err := suite.Client.DB.Integration.Create().
		SetName("Dup Convergence B").
		SetKind("dirduptest").
		SetDefinitionID("def_dirduptest").
		Save(ctx)
	th.RequireNoError(t, err)

	t.Cleanup(func() {
		for _, integrationID := range []string{one.ID, other.ID} {
			_, err := suite.Client.DB.DirectoryMembership.Delete().Where(directorymembership.IntegrationID(integrationID)).Exec(ctx)
			th.RequireNoError(t, err)
		}

		_, err := suite.Client.DB.DirectoryAccount.Delete().Where(directoryaccount.ExternalIDHasPrefix("dirdup-")).Exec(ctx)
		th.RequireNoError(t, err)
		_, err = suite.Client.DB.DirectoryGroup.Delete().Where(directorygroup.ExternalIDHasPrefix("dirdup-")).Exec(ctx)
		th.RequireNoError(t, err)

		for _, integrationID := range []string{one.ID, other.ID} {
			_, err := suite.Client.DB.DirectorySyncRun.Delete().Where(directorysyncrun.IntegrationID(integrationID)).Exec(ctx)
			th.RequireNoError(t, err)
		}

		th.RequireNoError(t, suite.Client.DB.Integration.DeleteOneID(one.ID).Exec(ctx))
		th.RequireNoError(t, suite.Client.DB.Integration.DeleteOneID(other.ID).Exec(ctx))
	})

	accounts := []string{`{"external_id":"dirdup-a-1","canonical_email":"dirdup1@example.com","display_name":"Dup User","profile":{"id":"dirdup-a-1","rev":1}}`}
	groups := []string{`{"external_id":"dirdup-g-1","display_name":"Dup Group","profile":{"id":"dirdup-g-1","rev":1}}`}
	memberships := []string{`{"directory_account_id":"dirdup-a-1","directory_group_id":"dirdup-g-1"}`}

	assert.Check(t, is.Equal(0, ingestDirectorySnapshot(ctx, t, one, accounts, groups, memberships).Failed))
	assert.Check(t, is.Equal(0, ingestDirectorySnapshot(ctx, t, other, accounts, groups, memberships).Failed))

	duplicates, err := suite.Client.DB.DirectoryAccount.Query().Where(directoryaccount.ExternalID("dirdup-a-1")).Count(ctx)
	th.RequireNoError(t, err)
	assert.Check(t, is.Equal(2, duplicates), "integration-scoped legacy syncs create one row per integration")

	// both installations resolve the same instance identity, and subsequent syncs converge
	one, err = suite.Client.DB.Integration.UpdateOneID(one.ID).
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "dup-tenant-1"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	other, err = suite.Client.DB.Integration.UpdateOneID(other.ID).
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "dup-tenant-1"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	assert.Check(t, is.Equal(0, ingestDirectorySnapshot(ctx, t, one, accounts, groups, memberships).Failed), "the first stamped sync must not fail on its own duplicate")
	assert.Check(t, is.Equal(0, ingestDirectorySnapshot(ctx, t, other, accounts, groups, memberships).Failed), "the second integration's stamped sync must converge on its own row, not fail")
	assert.Check(t, is.Equal(0, ingestDirectorySnapshot(ctx, t, one, accounts, groups, memberships).Failed), "syncs after both rows are stamped must keep converging without conflicts")

	remaining, err := suite.Client.DB.DirectoryAccount.Query().Where(directoryaccount.ExternalID("dirdup-a-1")).Count(ctx)
	th.RequireNoError(t, err)
	assert.Check(t, is.Equal(2, remaining), "unmerged duplicates persist without multiplying")

	active, err := suite.Client.DB.DirectoryMembership.Query().
		Where(directorymembership.IntegrationIDIn(one.ID, other.ID)).
		Where(directorymembership.RemovedAtIsNil()).
		Count(ctx)
	th.RequireNoError(t, err)
	assert.Check(t, is.Equal(1, active), "membership confirmations converge on one active shared row")
}
