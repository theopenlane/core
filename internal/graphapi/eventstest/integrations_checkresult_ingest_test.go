//go:build test

package eventstest_test

import (
	"context"
	"encoding/json"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/checkresult"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	intregistry "github.com/theopenlane/core/v2/internal/integrations/registry"
	integrationtypes "github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
)

const checkResultIngestTestOperation = "checkresult.ingest"

// checkResultIngestTestDefinition builds a minimal check-result ingest definition whose mapping passes provider payloads through unchanged
func checkResultIngestTestDefinition(defID string) integrationtypes.Definition {
	passthrough := integrationtypes.MappingOverride{MapExpr: "payload"}

	return integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{
			ID:          defID,
			DisplayName: "CheckResult Ingest Test",
			Active:      true,
		},
		Operations: []integrationtypes.OperationRegistration{
			{
				Name:  checkResultIngestTestOperation,
				Topic: gala.TopicName("integration." + defID + "." + checkResultIngestTestOperation),
				IngestHandle: func(context.Context, integrationtypes.OperationRequest) ([]integrationtypes.IngestPayloadSet, error) {
					return nil, nil
				},
				Ingest: []integrationtypes.IngestContract{{Schema: entityops.SchemaCheckResult.Name}},
			},
		},
		Mappings: []integrationtypes.MappingRegistration{{Schema: entityops.SchemaCheckResult.Name, Spec: passthrough}},
	}
}

// ingestCheckResultPayloads pushes check-result payloads through the synchronous catalog ingest path and returns the record-level result
func ingestCheckResultPayloads(ctx context.Context, t *testing.T, installation *ent.Integration, payloads ...string) operations.IngestResult {
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
	}, operations.IngestOptions{})
	th.RequireNoError(t, err)

	return result
}

// checkResultByParentExternalID loads one ingested check result by its lookup key
func checkResultByParentExternalID(ctx context.Context, t *testing.T, parentExternalID string) *ent.CheckResult {
	t.Helper()

	cr, err := suite.Client.DB.CheckResult.Query().Where(checkresult.ParentExternalID(parentExternalID)).Only(ctx)
	th.RequireNoError(t, err)

	return cr
}

// TestCheckResultReingestUpdatesInPlace verifies a second ingest with a changed field updates the row in place
func TestCheckResultReingestUpdatesInPlace(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	installation, err := suite.Client.DB.Integration.Create().
		SetName("CheckResult Ingest Test").
		SetKind("checkresulttest").
		SetDefinitionID("def_checkresulttest").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-checkresulttest"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	t.Cleanup(func() {
		results, err := suite.Client.DB.CheckResult.Query().Where(checkresult.ParentExternalIDHasPrefix("checkc-")).All(ctx)
		th.RequireNoError(t, err)

		if len(results) > 0 {
			(&th.Cleanup[*ent.CheckResultDeleteOne]{Client: suite.Client.DB.CheckResult, IDs: lo.Map(results, func(cr *ent.CheckResult, _ int) string { return cr.ID })}).MustDelete(th.SharedTestUser1.UserCtx, t)
		}

		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: installation.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	create := `{"parent_external_id":"checkc-1","source":"checkresult-test","details":"initial details"}`

	result := ingestCheckResultPayloads(ctx, t, installation, create)
	assert.Check(t, is.Equal(1, result.Persisted))
	assert.Check(t, is.Equal(0, result.Failed))
	assert.Check(t, is.Equal(1, result.Changed), "a created row must count as changed")

	changed := `{"parent_external_id":"checkc-1","source":"checkresult-test","details":"updated details"}`

	result = ingestCheckResultPayloads(ctx, t, installation, changed)
	assert.Check(t, is.Equal(1, result.Persisted), "a re-ingest with a changed field must update the row instead of failing")
	assert.Check(t, is.Equal(0, result.Failed), "a re-ingest of an existing record must not hit the cross-organization conflict guard")
	assert.Check(t, is.Equal(1, result.Changed))

	after := checkResultByParentExternalID(ctx, t, "checkc-1")
	assert.Check(t, is.Equal("updated details", lo.FromPtr(after.Details)))
}

// TestCheckResultFailedRecordExclusionAcrossBatchedRuns verifies a record that fails validation on
// every batched run is recorded and requeue-eligible on its first failure (counted Failed), then
// excluded on every subsequent batched run while tracked (counted Excluded, not Failed, attempts
// incrementing), and drops out of tracking once it reaches the retention ceiling
func TestCheckResultFailedRecordExclusionAcrossBatchedRuns(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	installation, err := suite.Client.DB.Integration.Create().
		SetName("CheckResult Exclusion Test").
		SetKind("checkresultexclusion").
		SetDefinitionID("def_checkresultexclusion").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-checkresultexclusion"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	t.Cleanup(func() {
		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: installation.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	// an out-of-range status fails CheckResult's enum validator on every attempt, deterministically,
	// without ever creating a row; parent_external_id stays present so the failed-record key resolves
	invalid := `{"parent_external_id":"checkexcl-1","status":"not-a-real-status","details":"invalid status"}`

	first := ingestCheckResultPayloads(ctx, t, installation, invalid)
	assert.Check(t, is.Equal(1, first.Failed), "the first failure of an untracked key must count as Failed")
	assert.Check(t, is.Equal(0, first.Excluded))

	reloaded, err := suite.Client.DB.Integration.Get(ctx, installation.ID)
	th.RequireNoError(t, err)
	assert.Check(t, is.Len(reloaded.Health.FailedRecords, 1))
	assert.Check(t, is.Equal(1, reloaded.Health.FailedRecords[0].Attempts))

	for attempt := 2; attempt <= 4; attempt++ {
		result := ingestCheckResultPayloads(ctx, t, installation, invalid)
		assert.Check(t, is.Equal(0, result.Failed), "a tracked failing key must not count as Failed once excluded")
		assert.Check(t, is.Equal(1, result.Excluded))

		reloaded, err = suite.Client.DB.Integration.Get(ctx, installation.ID)
		th.RequireNoError(t, err)
		assert.Check(t, is.Len(reloaded.Health.FailedRecords, 1))
		assert.Check(t, is.Equal(attempt, reloaded.Health.FailedRecords[0].Attempts))
	}

	fifth := ingestCheckResultPayloads(ctx, t, installation, invalid)
	assert.Check(t, is.Equal(0, fifth.Failed))
	assert.Check(t, is.Equal(1, fifth.Excluded))

	reloaded, err = suite.Client.DB.Integration.Get(ctx, installation.ID)
	th.RequireNoError(t, err)
	assert.Check(t, is.Len(reloaded.Health.FailedRecords, 0), "a key must drop out of tracking once it reaches the retention ceiling")
}
