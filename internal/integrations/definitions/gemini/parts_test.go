package gemini

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/internal/docextract/soc2"
)

func TestRequestedPartsEmpty(t *testing.T) {
	parts, err := RequestedParts(nil)

	assert.NilError(t, err)
	assert.Check(t, len(parts) == 0)

	parts, err = RequestedParts(map[string]any{requestedPartsMetadataKey: []any{}})

	assert.NilError(t, err)
	assert.Check(t, len(parts) == 0)
}

func TestRequestedParts(t *testing.T) {
	parts, err := RequestedParts(map[string]any{requestedPartsMetadataKey: []any{"controls", "reviews"}})

	assert.NilError(t, err)
	assert.DeepEqual(t, parts, []string{"controls", "reviews"})
}

func TestRequestedPartsUnknown(t *testing.T) {
	_, err := RequestedParts(map[string]any{requestedPartsMetadataKey: []any{"controls", "invoices"}})

	assert.ErrorIs(t, err, ErrUnknownRequestedPart)
	assert.ErrorContains(t, err, `"invoices"`)
}

func TestSelectPartsImportDefault(t *testing.T) {
	request := ReportScanRequest{parts: []string{"controls", "reviews"}}

	selected, err := request.selectParts(nil, soc2.ImportProfile)

	assert.NilError(t, err)
	assert.DeepEqual(t, selected, []string{"controls", "reviews"})
}

func TestSelectPartsImportLeavesOutAssessment(t *testing.T) {
	request := ReportScanRequest{parts: []string{"controls", "findings", "assessment"}}

	selected, err := request.selectParts(nil, soc2.ImportProfile)

	assert.NilError(t, err)
	assert.DeepEqual(t, selected, []string{"controls", "findings"})
}

func TestSelectPartsVendorReviewDefault(t *testing.T) {
	request := ReportScanRequest{parts: []string{"controls", "findings", "reviews", "assessment"}}

	selected, err := request.selectParts(nil, soc2.VendorReviewProfile)

	assert.NilError(t, err)
	assert.DeepEqual(t, selected, []string{"findings", "assessment"})
}

func TestSelectPartsVendorReviewMissingAssessment(t *testing.T) {
	request := ReportScanRequest{parts: []string{"controls", "findings"}}

	_, err := request.selectParts(nil, soc2.VendorReviewProfile)

	assert.ErrorIs(t, err, ErrRequestedPartsUnavailable)
}

func TestSelectPartsIntersectsWithConfigured(t *testing.T) {
	request := ReportScanRequest{parts: []string{"controls", "reviews", "assets"}}

	selected, err := request.selectParts(map[string]any{requestedPartsMetadataKey: []any{"reviews", "findings"}}, soc2.ImportProfile)

	assert.NilError(t, err)
	assert.DeepEqual(t, selected, []string{"reviews"})
}

func TestSelectPartsNoneAvailable(t *testing.T) {
	request := ReportScanRequest{parts: []string{"controls"}}

	_, err := request.selectParts(map[string]any{requestedPartsMetadataKey: []any{"findings"}}, soc2.ImportProfile)

	assert.ErrorIs(t, err, ErrRequestedPartsUnavailable)
}

func TestOrderPartsPutsControlsFirst(t *testing.T) {
	ordered := orderParts([]string{"assets", "reviews", "programs", "controls"})

	assert.DeepEqual(t, ordered, []string{"controls", "programs", "assets", "reviews"})
}

func TestPartStartAtReleasesWaves(t *testing.T) {
	now := time.Now()

	for index := range PartWaveSize {
		assert.Check(t, partStartAt(now, index).Equal(now))
	}

	assert.Check(t, partStartAt(now, PartWaveSize).Equal(now.Add(PartWaveInterval)))
	assert.Check(t, partStartAt(now, 2*PartWaveSize+1).Equal(now.Add(2*PartWaveInterval)))
}

func TestBackoffFromGrowsAndCaps(t *testing.T) {
	first := backoffFrom(PartQuotaMinInterval, 0)
	assert.Check(t, first >= PartQuotaMinInterval && first <= PartQuotaMinInterval+PartQuotaMinInterval/4)

	second := backoffFrom(PartQuotaMinInterval, 1)
	assert.Check(t, second >= 2*PartQuotaMinInterval)

	capped := backoffFrom(PartQuotaMinInterval, 20)
	assert.Check(t, capped >= PartMaxInterval && capped <= PartMaxInterval+PartMaxInterval/4)
}

func TestPartUniqueKeyChangesWithQuotaWaits(t *testing.T) {
	envelope := ReportScanPartEnvelope{ScanID: "scan", Part: "controls"}
	waited := envelope
	waited.QuotaWaits = 1

	assert.Check(t, reportScanPartTopic.UniqueKey(envelope) != reportScanPartTopic.UniqueKey(waited))
}
