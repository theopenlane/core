package soc2

import (
	"regexp"
	"slices"

	"github.com/theopenlane/core/v2/pkg/docextract"
)

const (
	// MinPages is the fewest pages a pdf can have and still be treated as a SOC 2 report
	MinPages = 5
	// MinStrongMarkers is how many strong markers must accompany the SOC 2 reference to accept the report
	MinStrongMarkers = 2
	// HighConfidenceMarkers is how many strong markers make the report a high-confidence match
	HighConfidenceMarkers = 4

	// ReportTypeOne is a SOC 2 Type 1 report, an as-of opinion with no testing results
	ReportTypeOne docextract.ReportType = "type1"
	// ReportTypeTwo is a SOC 2 Type 2 report, testing results over a reporting period
	ReportTypeTwo docextract.ReportType = "type2"
)

// init registers the profile so docextract.Validate recognizes SOC 2 reports
func init() {
	docextract.RegisterProfile(KindName, profile)
}

// profile recognizes a SOC 2 report by content common to every SOC 2 report; findings,
// exceptions, and qualifications never affect the outcome
var profile = docextract.Profile{
	Kind:                "SOC 2 report",
	MinPages:            MinPages,
	MinStrong:           MinStrongMarkers,
	HighConfidence:      HighConfidenceMarkers,
	Subtype:             subtype,
	MissingSectionsHint: "the service auditor's report, management's assertion, or the system description",
	Markers: []docextract.Marker{
		{Name: "SOC 2", Weight: docextract.MarkerRequired, Pattern: regexp.MustCompile(`(?i)\bSOC\s*2\b`)},
		{Name: "auditor_report", Weight: docextract.MarkerStrong, Pattern: regexp.MustCompile(`(?i)service auditor'?s report`)},
		{Name: "management_assertion", Weight: docextract.MarkerStrong, Pattern: regexp.MustCompile(`(?i)management'?s? assertion|assertion of management`)},
		{Name: "trust_services_criteria", Weight: docextract.MarkerStrong, Pattern: regexp.MustCompile(`(?i)trust services criteria|\bAICPA\b`)},
		{Name: "system_description", Weight: docextract.MarkerStrong, Pattern: regexp.MustCompile(`(?i)description of (the|its|[a-z]+'?s) system|system description`)},
		{Name: "criteria_ids", Weight: docextract.MarkerStrong, Pattern: regexp.MustCompile(`\b(CC|A|C|PI|P)\d{1,2}\.\d{1,2}\b`)},
		{Name: "type2_testing", Weight: docextract.MarkerStrong, Pattern: regexp.MustCompile(`(?i)tests? of controls|results of tests|exceptions? noted|tests? of operating effectiveness`)},
		{Name: "reporting_period", Weight: docextract.MarkerSupporting, Pattern: regexp.MustCompile(`(?i)for the period|throughout the period|from [a-z]+ \d{1,2}, \d{4},? (through|to) [a-z]+ \d{1,2}, \d{4}`)},
		{Name: "as_of_date", Weight: docextract.MarkerSupporting, Pattern: regexp.MustCompile(`(?i)\bas of [a-z]+ \d{1,2}, \d{4}`)},
		{Name: "trust_categories", Weight: docextract.MarkerSupporting, Pattern: regexp.MustCompile(`(?i)\b(availability|confidentiality|processing integrity|privacy)\b`)},
		{Name: "auditor_cpa", Weight: docextract.MarkerSupporting, Pattern: regexp.MustCompile(`(?i)\bCPAs?\b|certified public accountants?`)},
		{Name: "cuecs", Weight: docextract.MarkerSupporting, Pattern: regexp.MustCompile(`(?i)complementary user entity controls|\bCUECs?\b|subservice organizations?`)},
	},
}

// subtype infers Type 2 from testing language plus a reporting period, Type 1 from an as-of
// date without any testing language
func subtype(found map[string]bool) docextract.ReportType {
	switch {
	case found["type2_testing"] && found["reporting_period"]:
		return ReportTypeTwo
	case found["as_of_date"] && !found["type2_testing"]:
		return ReportTypeOne
	default:
		return docextract.ReportTypeUnknown
	}
}

// MinSectionItems is the fewest items a section every SOC 2 report populates can return before
// the pass is treated as having come back thin
const MinSectionItems = 5

// thinSections are the sections every SOC 2 report populates, so a near-empty result from one of
// them means the pass went wrong rather than the report being sparse
var thinSections = []string{ControlsPart, ReviewsPart}

// SectionTooThin reports whether an extracted section returned fewer items than the report should
// contain, and the minimum it was measured against. A request scoped to a handful of controls
// expects at least one item per control, capped at the floor
func SectionTooThin(part string, items, scoped int) (int, bool) {
	if !slices.Contains(thinSections, part) {
		return 0, false
	}

	minimum := MinSectionItems
	if scoped > 0 {
		minimum = min(minimum, scoped)
	}

	return minimum, items < minimum
}
