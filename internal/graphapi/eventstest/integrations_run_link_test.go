//go:build test

package eventstest_test

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/finding"
	"github.com/theopenlane/core/v2/internal/graphapi"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	intregistry "github.com/theopenlane/core/v2/internal/integrations/registry"
	integrationtypes "github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
)

// ingestFindingPayloadsWithOptions pushes finding payloads through the synchronous catalog ingest
// path with caller-supplied ingest options, extending ingestFindingPayloads with run correlation
func ingestFindingPayloadsWithOptions(ctx context.Context, t *testing.T, installation *ent.Integration, options operations.IngestOptions, payloads ...string) operations.IngestResult {
	t.Helper()

	def := findingIngestTestDefinition(installation.DefinitionID)
	reg := intregistry.New()
	th.RequireNoError(t, reg.Register(def))

	envelopes := lo.Map(payloads, func(p string, _ int) integrationtypes.MappingEnvelope {
		return integrationtypes.MappingEnvelope{Payload: json.RawMessage(p)}
	})

	result, err := operations.ProcessPayloadSets(ctx, operations.IngestContext{
		Registry:    reg,
		DB:          suite.Client.DB,
		Integration: installation,
	}, findingIngestTestOperation, def.Operations[0].Ingest, def.Operations[0].Policy, []integrationtypes.IngestPayloadSet{
		{Schema: entityops.SchemaFinding.Name, Envelopes: envelopes},
	}, options)
	th.RequireNoError(t, err)

	return result
}

// TestDirectorySnapshotIntegrationRunLink verifies a directory snapshot ingest stamps
// integration_run_id from IngestOptions.RunID on created rows, leaves it unwritten on a
// volatile-only re-sync under a new run, and repoints it once a material change rides along
func TestDirectorySnapshotIntegrationRunLink(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	const prefix = "runlink"

	installation, err := suite.Client.DB.Integration.Create().
		SetName("Run Link Test").
		SetKind("runlinktest").
		SetDefinitionID("def_dirsynctest").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-runlinktest"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	cleanupDirectoryPrefix(t, ctx, []string{installation.ID}, prefix)

	runOne, err := suite.Client.DB.IntegrationRun.Create().
		SetIntegrationID(installation.ID).
		SetOwnerID(installation.OwnerID).
		Save(ctx)
	th.RequireNoError(t, err)

	runTwo, err := suite.Client.DB.IntegrationRun.Create().
		SetIntegrationID(installation.ID).
		SetOwnerID(installation.OwnerID).
		Save(ctx)
	th.RequireNoError(t, err)

	t.Cleanup(func() {
		(&th.Cleanup[*ent.IntegrationRunDeleteOne]{Client: suite.Client.DB.IntegrationRun, IDs: []string{runOne.ID, runTwo.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	counters, teardown := directoryEventCounters(t)
	defer teardown()

	base := newDirectorySnapshot(prefix)

	seeded := ingestDirectorySnapshotFixture(ctx, t, installation, base, true, operations.IngestOptions{RunID: runOne.ID})
	waitForGala(t, counters.Runtime)
	assert.Check(t, is.Equal(0, seeded.Failed))
	assert.Check(t, is.Equal(len(base.Accounts)+len(base.Groups)+len(base.Memberships), seeded.Changed))

	for _, a := range base.Accounts {
		row := directoryAccountByExternalID(ctx, t, a.ExternalID)
		assert.Check(t, is.Equal(runOne.ID, row.IntegrationRunID), "a created account must carry the ingesting run's id")
	}
	for _, g := range base.Groups {
		row := directoryGroupByExternalID(ctx, t, g.ExternalID)
		assert.Check(t, is.Equal(runOne.ID, row.IntegrationRunID), "a created group must carry the ingesting run's id")
	}
	for _, m := range base.Memberships {
		row := directoryMembershipByExternalIDs(ctx, t, m.DirectoryAccountID, m.DirectoryGroupID)
		assert.Check(t, is.Equal(runOne.ID, row.IntegrationRunID), "a created membership must carry the ingesting run's id")
	}

	resynced := ingestDirectorySnapshotFixture(ctx, t, installation, base.identical(), true, operations.IngestOptions{RunID: runTwo.ID})
	waitForGala(t, counters.Runtime)
	assert.Check(t, is.Equal(0, resynced.Failed))
	assert.Check(t, is.Equal(0, resynced.Changed), "an idle resync under a new run id must not count as changed")
	assert.Check(t, is.Equal(int64(0), counters.AccountUpdates.Load()), "an idle resync must not emit an account update event even though the run id differs")
	assert.Check(t, is.Equal(int64(0), counters.GroupUpdates.Load()))
	assert.Check(t, is.Equal(int64(0), counters.MembershipUpdates.Load()))

	for _, a := range base.Accounts {
		row := directoryAccountByExternalID(ctx, t, a.ExternalID)
		assert.Check(t, is.Equal(runOne.ID, row.IntegrationRunID), "a volatile-only run id change must not be written")
	}
	for _, g := range base.Groups {
		row := directoryGroupByExternalID(ctx, t, g.ExternalID)
		assert.Check(t, is.Equal(runOne.ID, row.IntegrationRunID))
	}
	for _, m := range base.Memberships {
		row := directoryMembershipByExternalIDs(ctx, t, m.DirectoryAccountID, m.DirectoryGroupID)
		assert.Check(t, is.Equal(runOne.ID, row.IntegrationRunID))
	}

	changedExternalID := base.Accounts[0].ExternalID
	materialChange := base.withAccountMaterialChange(changedExternalID)

	thirdRun := ingestDirectorySnapshotFixture(ctx, t, installation, materialChange, true, operations.IngestOptions{RunID: runTwo.ID})
	waitForGala(t, counters.Runtime)
	assert.Check(t, is.Equal(0, thirdRun.Failed))
	assert.Check(t, is.Equal(1, thirdRun.Changed), "the material change must count as exactly one changed record")
	assert.Check(t, is.Equal(int64(1), counters.AccountUpdates.Load()), "the material change must emit exactly one account update event")

	changedRow := directoryAccountByExternalID(ctx, t, changedExternalID)
	assert.Check(t, is.Equal(runTwo.ID, changedRow.IntegrationRunID), "a material change must repoint the run id to the current run")

	for _, a := range base.Accounts {
		if a.ExternalID == changedExternalID {
			continue
		}

		row := directoryAccountByExternalID(ctx, t, a.ExternalID)
		assert.Check(t, is.Equal(runOne.ID, row.IntegrationRunID), "an untouched account must keep its original run id")
	}
	for _, g := range base.Groups {
		row := directoryGroupByExternalID(ctx, t, g.ExternalID)
		assert.Check(t, is.Equal(runOne.ID, row.IntegrationRunID), "an untouched group must keep its original run id")
	}
	for _, m := range base.Memberships {
		row := directoryMembershipByExternalIDs(ctx, t, m.DirectoryAccountID, m.DirectoryGroupID)
		assert.Check(t, is.Equal(runOne.ID, row.IntegrationRunID), "an untouched membership must keep its original run id")
	}
}

// TestFindingIntegrationRunLink verifies a Finding ingest stamps integration_run_id from
// IngestOptions.RunID on a created row, leaves it unwritten on a volatile-only re-ingest under a
// new run, and repoints it once a material change rides along
func TestFindingIntegrationRunLink(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	installation, err := suite.Client.DB.Integration.Create().
		SetName("Finding Run Link Test").
		SetKind("findrunlinktest").
		SetDefinitionID("def_findrunlink").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-findrunlinktest"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	runOne, err := suite.Client.DB.IntegrationRun.Create().
		SetIntegrationID(installation.ID).
		SetOwnerID(installation.OwnerID).
		Save(ctx)
	th.RequireNoError(t, err)

	runTwo, err := suite.Client.DB.IntegrationRun.Create().
		SetIntegrationID(installation.ID).
		SetOwnerID(installation.OwnerID).
		Save(ctx)
	th.RequireNoError(t, err)

	var findingUpdates atomic.Int64

	setup, err := graphapi.SetupListenerRuntime(suite.GalaRuntime, []gala.Registration{
		entityops.MutationListener{
			Schema:     entityops.SchemaFinding,
			Operations: []string{entityops.OpUpdateOne},
			Handle: func(_ entityops.Invocation, _ entityops.MutationPayload) error {
				findingUpdates.Add(1)

				return nil
			},
		},
	})
	assert.NilError(t, err)
	defer setup.Teardown()

	t.Cleanup(func() {
		findings, err := suite.Client.DB.Finding.Query().Where(finding.ExternalIDHasPrefix("find-runlink-")).All(ctx)
		th.RequireNoError(t, err)

		if len(findings) > 0 {
			(&th.Cleanup[*ent.FindingDeleteOne]{Client: suite.Client.DB.Finding, IDs: lo.Map(findings, func(f *ent.Finding, _ int) string { return f.ID })}).MustDelete(th.SharedTestUser1.UserCtx, t)
		}

		(&th.Cleanup[*ent.IntegrationRunDeleteOne]{Client: suite.Client.DB.IntegrationRun, IDs: []string{runOne.ID, runTwo.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: installation.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	create := `{"external_id":"find-runlink-1","display_name":"Run Link Finding","description":"initial desc"}`

	result := ingestFindingPayloadsWithOptions(ctx, t, installation, operations.IngestOptions{RunID: runOne.ID}, create)
	assert.Check(t, is.Equal(1, result.Changed), "the create must count as changed")

	created := findingByExternalID(ctx, t, "find-runlink-1")
	assert.Check(t, is.Equal(runOne.ID, created.IntegrationRunID), "a created finding must carry the ingesting run's id")

	result = ingestFindingPayloadsWithOptions(ctx, t, installation, operations.IngestOptions{RunID: runTwo.ID}, create)
	waitForGala(t, setup.Runtime)

	assert.Check(t, is.Equal(0, result.Changed), "an identical re-ingest under a new run id must not count as changed")
	assert.Check(t, is.Equal(int64(0), findingUpdates.Load()), "a volatile-only run id change must not emit an update event")

	unchanged := findingByExternalID(ctx, t, "find-runlink-1")
	assert.Check(t, is.Equal(runOne.ID, unchanged.IntegrationRunID), "a volatile-only run id change must not be written")
	assert.Check(t, unchanged.UpdatedAt.Equal(created.UpdatedAt))

	materialChange := `{"external_id":"find-runlink-1","display_name":"Run Link Finding","description":"changed desc"}`

	result = ingestFindingPayloadsWithOptions(ctx, t, installation, operations.IngestOptions{RunID: runTwo.ID}, materialChange)
	waitForGala(t, setup.Runtime)

	assert.Check(t, is.Equal(1, result.Changed), "the material change must count as changed")
	assert.Check(t, is.Equal(int64(1), findingUpdates.Load()), "the material change must emit exactly one update event")

	after := findingByExternalID(ctx, t, "find-runlink-1")
	assert.Check(t, is.Equal("changed desc", after.Description))
	assert.Check(t, is.Equal(runTwo.ID, after.IntegrationRunID), "the material change must repoint the run id to the current run")
}
