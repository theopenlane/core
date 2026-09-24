package soc2

import (
	"context"
	"encoding/json"
	"regexp"
	"slices"
	"strings"

	"github.com/theopenlane/core/v2/pkg/docextract/schema"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// criterionPattern matches a SOC 2 criterion code such as CC6.1 or A1.2, as opposed to a
// control identifier printed by the report such as ACF-60 or a derived code such as CC6.1.1
var criterionPattern = regexp.MustCompile(`^[A-Z]+\d+\.\d+$`)

// reviewStream merges streamed review payloads by external id and tracks which controls have
// been covered so a continuation can resume after the last one
type reviewStream struct {
	prompts  Prompts
	controls map[string][]string
	ordered  []string
}

// newReviewStream starts tracking for one reviews extraction
func newReviewStream(prompts Prompts) *reviewStream {
	return &reviewStream{prompts: prompts, controls: map[string][]string{}}
}

// Merge folds the reviews in streamed into collected, keyed by external id and sorted for stable output
func (*reviewStream) Merge(streamed, collected string) (string, error) {
	var incoming schema.Reviews
	if err := json.Unmarshal([]byte(streamed), &incoming); err != nil {
		return "", err
	}

	existing := schema.Reviews{Reviews: []schema.Review{}}

	if collected != "" {
		if err := json.Unmarshal([]byte(collected), &existing); err != nil {
			return "", err
		}
	}

	byExternalID := make(map[string]schema.Review, len(existing.Reviews)+len(incoming.Reviews))
	for _, review := range existing.Reviews {
		byExternalID[review.ExternalID] = review
	}

	for _, review := range incoming.Reviews {
		byExternalID[review.ExternalID] = review
	}

	combined := schema.Reviews{Reviews: make([]schema.Review, 0, len(byExternalID))}
	for _, review := range byExternalID {
		combined.Reviews = append(combined.Reviews, review)
	}

	slices.SortFunc(combined.Reviews, compareReviewExternalID)

	merged, err := json.Marshal(combined)
	if err != nil {
		return "", err
	}

	return string(merged), nil
}

// Count returns how many reviews the output holds
func (*reviewStream) Count(output string) int {
	return ReviewCount(output)
}

// Continuation records the controls covered by the streamed payload and renders the prompt that
// asks the model to resume after the last one
func (s *reviewStream) Continuation(ctx context.Context, streamed string) (string, error) {
	lastControl, lastCriterion := s.track(ctx, streamed)

	logx.FromContext(ctx).Debug().Int("controls_covered", len(s.ordered)).Str("continue_after_control", lastControl).Str("continue_after_criterion", lastCriterion).Msg("soc2: requesting continuation")

	return s.prompts.continuationPrompt(lastControl, lastCriterion)
}

// track folds the reviews in streamed into the control map, keeping control order, and returns
// the last control seen with its first criterion
func (s *reviewStream) track(ctx context.Context, streamed string) (string, string) {
	var reviews schema.Reviews

	if err := json.Unmarshal([]byte(streamed), &reviews); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("soc2: failed to unmarshal reviews for control tracking")

		return "", ""
	}

	lastControl := ""

	for _, review := range reviews.Reviews {
		controlCode, criteria := splitRefCodes(review.RefCodes)
		if controlCode == "" {
			continue
		}

		lastControl = controlCode

		if _, exists := s.controls[controlCode]; !exists {
			s.controls[controlCode] = []string{}
			s.ordered = append(s.ordered, controlCode)
		}

		for _, criterion := range criteria {
			if !slices.Contains(s.controls[controlCode], criterion) {
				s.controls[controlCode] = append(s.controls[controlCode], criterion)
			}
		}
	}

	if lastControl == "" {
		return "", ""
	}

	lastCriterion := ""

	if len(s.controls[lastControl]) > 0 {
		lastCriterion = s.controls[lastControl][0]
	}

	return lastControl, lastCriterion
}

// splitRefCodes separates a review's ref codes into the control code and the criteria it maps
// to; the control code is the first non-criterion entry, falling back to the first criterion
// when the report printed no control identifier
func splitRefCodes(refCodes []string) (string, []string) {
	controlCode := ""
	criteria := make([]string, 0, len(refCodes))

	for _, refCode := range refCodes {
		if criterionPattern.MatchString(refCode) {
			criteria = append(criteria, refCode)

			continue
		}

		if controlCode == "" {
			controlCode = refCode
		}
	}

	if controlCode == "" && len(criteria) > 0 {
		controlCode = criteria[0]
	}

	return controlCode, criteria
}

// ReviewCount returns how many reviews a reviews payload holds, zero when it is empty or malformed
func ReviewCount(output string) int {
	if output == "" {
		return 0
	}

	var reviews schema.Reviews
	if err := json.Unmarshal([]byte(output), &reviews); err != nil {
		return 0
	}

	return len(reviews.Reviews)
}

// compareReviewExternalID orders reviews by external id so merged output is stable
func compareReviewExternalID(a, b schema.Review) int {
	return strings.Compare(a.ExternalID, b.ExternalID)
}
