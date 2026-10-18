package soc2

import (
	"context"
	"regexp"
	"slices"
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/pkg/docextract"
	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/core/v2/pkg/mapx"
)

// controlSection describes how the items of a control-scoped section are identified
type controlSection struct {
	// textFields hold the test text an item restates, together identifying it within its control
	textFields []string
	// externalIDField identifies an item that carries no test text
	externalIDField string
}

// controlSections are the control-scoped sections that stream, keyed by section name
var controlSections = map[string]controlSection{
	ReviewsPart:  {textFields: []string{entityops.ReviewFields.Details.Name}, externalIDField: entityops.ReviewFields.ExternalID.Name},
	FindingsPart: {textFields: []string{testDetailsKey, entityops.FindingFields.Description.Name}, externalIDField: entityops.FindingFields.ExternalID.Name},
}

// criterionPattern matches a SOC 2 criterion code such as CC6.1 or A1.2, as opposed to a
// control identifier printed by the report such as ACF-60 or a derived code such as CC6.1.1
var criterionPattern = regexp.MustCompile(`^[A-Z]+\d+\.\d+$`)

// controlStream identifies and folds the items of one control-scoped section and tracks which
// controls have been covered so a continuation can resume after the last one
type controlStream struct {
	section  string
	config   controlSection
	prompts  Prompts
	controls map[string][]string
	ordered  []string
}

// newControlStream starts tracking for one control-scoped section extraction
func newControlStream(section string, prompts Prompts) *controlStream {
	return &controlStream{section: section, config: controlSections[section], prompts: prompts, controls: map[string][]string{}}
}

// stream returns the section stream that merges and continues this section
func (s *controlStream) stream() *docextract.SectionStream {
	return &docextract.SectionStream{
		Section:   s.section,
		Key:       s.key,
		Fold:      foldRefCodes,
		SortField: s.config.externalIDField,
		Resume:    s.resume,
	}
}

// key identifies the test an item records within its control, falling back to its external id when
// it carries no test text; a report prints the same test once per criterion it satisfies, which
// would otherwise land as several items
func (s *controlStream) key(item map[string]any) (string, bool) {
	controlCode, _ := splitRefCodes(refCodesOf(item))

	text := strings.Join(lo.Compact(lo.Map(s.config.textFields, func(field string, _ int) string { return mapx.Value[string](item, field) })), "\x00")

	if key, ok := testKey(controlCode, text); ok {
		return key, true
	}

	externalID := mapx.Value[string](item, s.config.externalIDField)

	return externalID, externalID != ""
}

// foldRefCodes keeps the latest item, collecting every ref code the report restated it under in the order seen
func foldRefCodes(previous, item map[string]any) map[string]any {
	item[refCodesKey] = lo.ToAnySlice(lo.Union(refCodesOf(previous), refCodesOf(item)))

	return item
}

// resume records the controls covered by the streamed items and renders the prompt that asks the
// model to resume after the last one
func (s *controlStream) resume(ctx context.Context, items []map[string]any) (string, error) {
	lastControl, lastCriterion := s.track(items)

	logx.FromContext(ctx).Debug().Str("section", s.section).Int("controls_covered", len(s.ordered)).Str("continue_after_control", lastControl).Str("continue_after_criterion", lastCriterion).Msg("soc2: requesting continuation")

	return s.prompts.continuationPrompt(s.section, lastControl, lastCriterion)
}

// track folds the streamed items into the control map, keeping control order, and returns
// the last control seen with its first criterion
func (s *controlStream) track(items []map[string]any) (string, string) {
	lastControl := ""

	for _, item := range items {
		controlCode, criteria := splitRefCodes(refCodesOf(item))
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
