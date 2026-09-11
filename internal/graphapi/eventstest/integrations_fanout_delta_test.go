//go:build test

package eventstest_test

import (
	"testing"
	"time"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	openapi "github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/finding"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
)

const fanoutDeltaTestOperation = "fanout.delta.ingest"

// TestFanoutDeltaLinkedRecordCount verifies LastSuccessfulRunID resolves the most recent successful
// run for an installation and operation while skipping a still-running one, and that
// LinkedRecordCount sums the ingested records carrying that run's id, the pair the fan-out reconcile
// branch composes to estimate its adaptive-scheduling delta
func TestFanoutDeltaLinkedRecordCount(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	installation, err := suite.Client.DB.Integration.Create().
		SetName("Fanout Delta Test").
		SetKind("fanoutdeltatest").
		SetDefinitionID("def_fanoutdelta").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-fanoutdeltatest"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	noPriorRunID, err := operations.LastSuccessfulRunID(ctx, suite.Client.DB, installation.ID, fanoutDeltaTestOperation)
	th.RequireNoError(t, err)
	assert.Check(t, is.Equal("", noPriorRunID), "an installation with no completed run has no predecessor")

	runOne, err := suite.Client.DB.IntegrationRun.Create().
		SetIntegrationID(installation.ID).
		SetOwnerID(installation.OwnerID).
		SetOperationName(fanoutDeltaTestOperation).
		SetStatus(enums.IntegrationRunStatusSuccess).
		SetFinishedAt(time.Now().Add(-time.Hour)).
		Save(ctx)
	th.RequireNoError(t, err)

	runTwo, err := suite.Client.DB.IntegrationRun.Create().
		SetIntegrationID(installation.ID).
		SetOwnerID(installation.OwnerID).
		SetOperationName(fanoutDeltaTestOperation).
		SetStatus(enums.IntegrationRunStatusRunning).
		Save(ctx)
	th.RequireNoError(t, err)

	t.Cleanup(func() {
		findings, err := suite.Client.DB.Finding.Query().Where(finding.ExternalIDHasPrefix("find-fanoutdelta-")).All(ctx)
		th.RequireNoError(t, err)

		if len(findings) > 0 {
			(&th.Cleanup[*ent.FindingDeleteOne]{Client: suite.Client.DB.Finding, IDs: lo.Map(findings, func(f *ent.Finding, _ int) string { return f.ID })}).MustDelete(th.SharedTestUser1.UserCtx, t)
		}

		(&th.Cleanup[*ent.IntegrationRunDeleteOne]{Client: suite.Client.DB.IntegrationRun, IDs: []string{runOne.ID, runTwo.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: installation.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	resolvedRunID, err := operations.LastSuccessfulRunID(ctx, suite.Client.DB, installation.ID, fanoutDeltaTestOperation)
	th.RequireNoError(t, err)
	assert.Check(t, is.Equal(runOne.ID, resolvedRunID), "the most recent successful run must be resolved, not the still-running one")

	zeroLinked, err := operations.LinkedRecordCount(ctx, suite.Client.DB, installation.OwnerID, resolvedRunID)
	th.RequireNoError(t, err)
	assert.Check(t, is.Equal(0, zeroLinked), "a run with no ingested records links zero rows")

	create1 := `{"external_id":"find-fanoutdelta-1","display_name":"Fanout Delta Finding 1"}`
	create2 := `{"external_id":"find-fanoutdelta-2","display_name":"Fanout Delta Finding 2"}`

	result := ingestFindingPayloadsWithOptions(ctx, t, installation, operations.IngestOptions{RunID: runOne.ID}, create1, create2)
	assert.Check(t, is.Equal(2, result.Changed), "both creates must count as changed")

	linked, err := operations.LinkedRecordCount(ctx, suite.Client.DB, installation.OwnerID, runOne.ID)
	th.RequireNoError(t, err)
	assert.Check(t, is.Equal(2, linked), "LinkedRecordCount must sum the records the previous run ingested")

	unlinkedToRunTwo, err := operations.LinkedRecordCount(ctx, suite.Client.DB, installation.OwnerID, runTwo.ID)
	th.RequireNoError(t, err)
	assert.Check(t, is.Equal(0, unlinkedToRunTwo), "a different run id must not pick up another run's records")
}
