//go:build test

package eventstest_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/checkresult"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	intregistry "github.com/theopenlane/core/v2/internal/integrations/registry"
	integrationtypes "github.com/theopenlane/core/v2/internal/integrations/types"
)

// ingestCheckResultPayloadsWithRunID pushes check-result payloads through the synchronous catalog
// ingest path tagged with the given integration run id, mirroring ingestCheckResultPayloads with an
// explicit run id so a tracked failure records the run that first observed it
func ingestCheckResultPayloadsWithRunID(ctx context.Context, t *testing.T, installation *ent.Integration, runID string, payloads ...string) operations.IngestResult {
	t.Helper()

	def := checkResultIngestTestDefinition(installation.DefinitionID)
	reg := intregistry.New()
	th.RequireNoError(t, reg.Register(def))

	envelopes := lo.Map(payloads, func(p string, _ int) integrationtypes.MappingEnvelope {
		return integrationtypes.MappingEnvelope{Payload: json.RawMessage(p)}
	})

	result, err := operations.ProcessPayloadSets(ctx, operations.IngestContext{
		Registry:    reg,
		DB:          suite.Client.DB,
		Integration: installation,
	}, checkResultIngestTestOperation, def.Operations[0].Ingest, def.Operations[0].Policy, []integrationtypes.IngestPayloadSet{
		{Schema: entityops.SchemaCheckResult.Name, Envelopes: envelopes},
	}, operations.IngestOptions{RunID: runID})
	th.RequireNoError(t, err)

	return result
}

// seedCheckResultDurableRow writes a check-result row directly through the ent client, simulating a
// durable per-record retry job that succeeded outside the batched ingest path: the row carries the
// installation's provenance and the given integration run id
func seedCheckResultDurableRow(ctx context.Context, t *testing.T, installation *ent.Integration, parentExternalID, runID, details string) *ent.CheckResult {
	t.Helper()

	row, err := suite.Client.DB.CheckResult.Create().
		SetParentExternalID(parentExternalID).
		SetSource("checkresult-retry-test").
		SetStatus(enums.CheckStatusPass).
		SetDetails(details).
		SetSourceDefinitionID(installation.DefinitionID).
		SetSourceInstanceID(installation.InstallationMetadata.Display.ExternalID).
		SetManagedBy(installation.ID).
		SetIntegrationID(installation.ID).
		SetIntegrationRunID(runID).
		Save(ctx)
	th.RequireNoError(t, err)

	return row
}

// trackedFailedRecord finds one key's tracked failure on the installation's health, if any
func trackedFailedRecord(health models.IntegrationHealth, key string) (models.FailedRecord, bool) {
	return lo.Find(health.FailedRecords, func(fr models.FailedRecord) bool { return fr.Key == key })
}

// TestCheckResultTrackedFailureResolvedByDurableRetry verifies a tracked failed record is released
// from exclusion once a durable per-record retry has independently written the row at or after the
// tracked run, and stays excluded when the row on hand predates the tracked run
func TestCheckResultTrackedFailureResolvedByDurableRetry(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	installation, err := suite.Client.DB.Integration.Create().
		SetName("CheckResult Retry Resolution Test").
		SetKind("checkresultretry").
		SetDefinitionID("def_checkresultretry").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-checkresultretry"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	t.Cleanup(func() {
		results, err := suite.Client.DB.CheckResult.Query().Where(checkresult.ParentExternalIDHasPrefix("checkretry-")).All(ctx)
		th.RequireNoError(t, err)

		if len(results) > 0 {
			(&th.Cleanup[*ent.CheckResultDeleteOne]{Client: suite.Client.DB.CheckResult, IDs: lo.Map(results, func(cr *ent.CheckResult, _ int) string { return cr.ID })}).MustDelete(th.SharedTestUser1.UserCtx, t)
		}

		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: installation.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	// runs are created in order so their ids sort stale < first < second: stale is the durable
	// row's run in the unresolved case, first records each tracked failure, second re-attempts
	// each tracked record with a material change
	staleRun, err := suite.Client.DB.IntegrationRun.Create().
		SetIntegrationID(installation.ID).
		SetOwnerID(installation.OwnerID).
		Save(ctx)
	th.RequireNoError(t, err)

	firstRun, err := suite.Client.DB.IntegrationRun.Create().
		SetIntegrationID(installation.ID).
		SetOwnerID(installation.OwnerID).
		Save(ctx)
	th.RequireNoError(t, err)

	secondRun, err := suite.Client.DB.IntegrationRun.Create().
		SetIntegrationID(installation.ID).
		SetOwnerID(installation.OwnerID).
		Save(ctx)
	th.RequireNoError(t, err)

	t.Cleanup(func() {
		(&th.Cleanup[*ent.IntegrationRunDeleteOne]{Client: suite.Client.DB.IntegrationRun, IDs: []string{staleRun.ID, firstRun.ID, secondRun.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	staleRunID, firstRunID, secondRunID := staleRun.ID, firstRun.ID, secondRun.ID

	// --- resolved case: a durable retry wrote the row at the tracked run id ---

	const resolvedKey = "checkretry-resolved"

	invalidResolved := `{"parent_external_id":"checkretry-resolved","status":"not-a-real-status","details":"invalid status"}`

	first := ingestCheckResultPayloadsWithRunID(ctx, t, installation, firstRunID, invalidResolved)
	assert.Check(t, is.Equal(1, first.Failed), "the first failure of an untracked key must count as Failed")
	assert.Check(t, is.Equal(0, first.Excluded))

	reloaded, err := suite.Client.DB.Integration.Get(ctx, installation.ID)
	th.RequireNoError(t, err)
	trackedEntry, ok := trackedFailedRecord(reloaded.Health, resolvedKey)
	assert.Check(t, ok, "the failing record must be tracked")
	assert.Check(t, is.Equal(firstRunID, trackedEntry.RunID))

	// simulate a durable per-record retry succeeding after the batched run recorded the failure:
	// the row is written directly with the tracked run id and the installation's provenance
	seedCheckResultDurableRow(ctx, t, installation, resolvedKey, firstRunID, "seed details")

	changedResolved := `{"parent_external_id":"checkretry-resolved","source":"checkresult-retry-test","status":"PASS","details":"resolved details"}`

	second := ingestCheckResultPayloadsWithRunID(ctx, t, installation, secondRunID, changedResolved)
	assert.Check(t, is.Equal(0, second.Excluded), "a tracked failure resolved by a durable retry must not be excluded")
	assert.Check(t, is.Equal(1, second.Changed), "the fallthrough handle() call must persist the material change")
	assert.Check(t, is.Equal(0, second.Failed))

	after := checkResultByParentExternalID(ctx, t, resolvedKey)
	assert.Check(t, is.Equal("resolved details", lo.FromPtr(after.Details)))

	reloaded, err = suite.Client.DB.Integration.Get(ctx, installation.ID)
	th.RequireNoError(t, err)
	_, stillTracked := trackedFailedRecord(reloaded.Health, resolvedKey)
	assert.Check(t, !stillTracked, "the resolved key must be dropped from tracking")

	// --- unresolved case: the row on hand predates the tracked run id ---

	const staleKey = "checkretry-stale"

	invalidStale := `{"parent_external_id":"checkretry-stale","status":"not-a-real-status","details":"invalid status"}`

	firstStale := ingestCheckResultPayloadsWithRunID(ctx, t, installation, firstRunID, invalidStale)
	assert.Check(t, is.Equal(1, firstStale.Failed))
	assert.Check(t, is.Equal(0, firstStale.Excluded))

	seedCheckResultDurableRow(ctx, t, installation, staleKey, staleRunID, "stale seed details")

	changedStale := `{"parent_external_id":"checkretry-stale","source":"checkresult-retry-test","status":"PASS","details":"attempted change"}`

	secondStale := ingestCheckResultPayloadsWithRunID(ctx, t, installation, secondRunID, changedStale)
	assert.Check(t, is.Equal(1, secondStale.Excluded), "a row predating the tracked run must not resolve the tracked failure")
	assert.Check(t, is.Equal(0, secondStale.Changed))
	assert.Check(t, is.Equal(0, secondStale.Failed))

	staleAfter := checkResultByParentExternalID(ctx, t, staleKey)
	assert.Check(t, is.Equal("stale seed details", lo.FromPtr(staleAfter.Details)), "an excluded record must never reach handle()")

	reloaded, err = suite.Client.DB.Integration.Get(ctx, installation.ID)
	th.RequireNoError(t, err)
	staleEntry, staleTracked := trackedFailedRecord(reloaded.Health, staleKey)
	assert.Check(t, staleTracked, "the stale key must remain tracked")
	assert.Check(t, is.Equal(2, staleEntry.Attempts))
}
