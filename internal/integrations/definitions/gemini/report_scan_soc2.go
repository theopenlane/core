package gemini

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/theopenlane/precis/docextract"
	"github.com/theopenlane/precis/docextract/kinds/soc2"

	"github.com/theopenlane/core/common/enums"

	"github.com/theopenlane/core/v2/internal/docextract/soc2schema"
	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/utils/mapx"
)

// This file holds what the report scan saga needs to know about SOC 2 reports. What a SOC 2
// report contains lives in precis/docextract/kinds/soc2; what stays here is how this integration batches
// and requests those sections

const (
	// ReviewBatchSize is how many controls one reviews job covers; every batch is one concurrent
	// model call, so a smaller batch trades quota pressure for shorter requests
	ReviewBatchSize = 20
	// FindingBatchSize is how many controls one findings job covers; most controls have none so batches can be larger
	FindingBatchSize = 50
)

// controlScopedBatchSizes are the sections scoped by control ref codes, with how many controls
// one job of each covers
var controlScopedBatchSizes = map[string]int{
	soc2.ReviewsPart:  ReviewBatchSize,
	soc2.FindingsPart: FindingBatchSize,
}

// thinSectionError returns an error when the section holds fewer items than a report of this kind
// should contain, so the pass can be retried while attempts remain
func thinSectionError(envelope ReportScanPartEnvelope, section json.RawMessage) error {
	var items []json.RawMessage
	if err := json.Unmarshal(section, &items); err != nil {
		return nil
	}

	scoped := 0
	if envelope.isBatch() {
		scoped = len(envelope.RefCodes)
	}

	minimum, thin := soc2.SectionTooThin(envelope.Part, len(items), scoped)
	if !thin {
		return nil
	}

	return fmt.Errorf("%w: %s returned %d item(s), expected at least %d", ErrSectionTooThin, envelope.Part, len(items), minimum)
}

// applyAssessmentOpinion records the validated opinion on a vendor assessment and raises its risk
// rating to the floor that opinion sets, so the model can never rate a report below its opinion
func applyAssessmentOpinion(ctx context.Context, metadata map[string]any) {
	report, _ := metadata[reportMetadataKey].(map[string]any)

	assessment, ok := report[soc2.AssessmentPart].(map[string]any)
	if !ok {
		return
	}

	opinion := docextract.OpinionUnknown

	if validation, ok := metadata[validationMetadataKey].(map[string]any); ok {
		if value, ok := validation["opinion"].(string); ok && value != "" {
			opinion = docextract.Opinion(value)
		}
	}

	rating := *enums.ToVendorRiskRating(mapx.Value[string](assessment, soc2schema.RiskRatingKey))
	floored := soc2schema.ApplyRiskFloor(rating, opinion)

	if floored != rating {
		logx.FromContext(ctx).Info().Str("model_rating", rating.String()).Str("rating", floored.String()).Str("opinion", opinion.String()).Msg("report scan: assessment risk raised to the opinion floor")
	}

	assessment["opinion"] = opinion.String()
	assessment[soc2schema.RiskRatingKey] = floored.String()
}
