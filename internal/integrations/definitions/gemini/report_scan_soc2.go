package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/theopenlane/core/v2/pkg/docextract"
	"github.com/theopenlane/core/v2/pkg/docextract/soc2"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// This file holds what the report scan saga needs to know about SOC 2 reports. What a SOC 2
// report contains lives in pkg/docextract/soc2; what stays here is how this integration batches
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

// RequestedParts reads the sections a scan asked for from its metadata, validating every name;
// an absent or empty list means every configured section
func RequestedParts(metadata map[string]any) ([]string, error) {
	raw, _ := metadata[requestedPartsMetadataKey].([]any)

	names := make([]string, 0, len(raw))

	for _, value := range raw {
		name, _ := value.(string)
		if _, ok := soc2.PartByName(name); !ok {
			return nil, fmt.Errorf("%w: %q, expected one of %s", ErrUnknownRequestedPart, name, strings.Join(soc2.PartNames(), ", "))
		}

		names = append(names, name)
	}

	return names, nil
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

	rating, _ := assessment["riskRating"].(string)
	floored := soc2.ApplyRiskFloor(rating, opinion)

	if floored != rating {
		logx.FromContext(ctx).Info().Str("model_rating", rating).Str("rating", floored).Str("opinion", opinion.String()).Msg("report scan: assessment risk raised to the opinion floor")
	}

	assessment["opinion"] = opinion.String()
	assessment["riskRating"] = floored
}
