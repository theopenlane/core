package gemini

import (
	"context"
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/enums"
)

func assessmentMetadata(rating, opinion string) map[string]any {
	return map[string]any{
		reportMetadataKey: map[string]any{
			"assessment": map[string]any{"summary": "Meow Platform, security availability confidentiality", "risk_rating": rating, "risk_rationale": "clean opinion", "vendor": map[string]any{"name": "Meow Cloud", "soc2_period_end": "2026-07-28"}},
		},
		validationMetadataKey: map[string]any{"opinion": opinion},
	}
}

func TestApplyAssessmentOpinionRaisesToFloor(t *testing.T) {
	metadata := assessmentMetadata("LOW", "qualified")

	applyAssessmentOpinion(context.Background(), metadata)

	assessment, ok := Assessment(metadata)

	assert.Assert(t, ok)
	assert.Check(t, assessment.RiskRating == enums.VendorRiskRatingMedium)
	assert.Check(t, assessment.Opinion == "qualified")
	assert.Check(t, assessment.Vendor.Soc2PeriodEnd == "2026-07-28")
}

func TestApplyAssessmentOpinionKeepsHigherRating(t *testing.T) {
	metadata := assessmentMetadata("HIGH", "unqualified")

	applyAssessmentOpinion(context.Background(), metadata)

	assessment, ok := Assessment(metadata)

	assert.Assert(t, ok)
	assert.Check(t, assessment.RiskRating == enums.VendorRiskRatingHigh)
	assert.Check(t, assessment.Opinion == "unqualified")
}

func TestApplyAssessmentOpinionWithoutAssessment(t *testing.T) {
	metadata := map[string]any{reportMetadataKey: map[string]any{"findings": []any{}}}

	applyAssessmentOpinion(context.Background(), metadata)

	_, ok := Assessment(metadata)

	assert.Check(t, !ok)
}

func TestSectionItemCountObject(t *testing.T) {
	assert.Check(t, sectionItemCount(json.RawMessage(`{"summary":"x"}`)) == 1)
	assert.Check(t, sectionItemCount(json.RawMessage(`[{},{}]`)) == 2)
	assert.Check(t, sectionItemCount(json.RawMessage(`{}`)) == 0)
}
