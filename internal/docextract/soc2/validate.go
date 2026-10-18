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

	// OpinionUnqualified is a clean opinion, the controls were suitably designed (and operating effectively for Type 2)
	OpinionUnqualified docextract.Opinion = "unqualified"
	// OpinionQualified is a clean opinion except for the matters described in a basis for qualified opinion
	OpinionQualified docextract.Opinion = "qualified"
	// OpinionAdverse means the description or controls are materially misstated or ineffective
	OpinionAdverse docextract.Opinion = "adverse"
	// OpinionDisclaimer means the auditor could not obtain enough evidence to issue an opinion
	OpinionDisclaimer docextract.Opinion = "disclaimer"
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
	Opinion:             opinion,
	MissingSectionsHint: "the service auditor's report, management's assertion, or the system description",
	Markers: []docextract.Marker{
		{Name: "SOC 2", Weight: docextract.MarkerRequired, Pattern: regexp.MustCompile(`(?i)\bSOC\W{0,3}2\b`)},
		{Name: "auditor_report", Weight: docextract.MarkerStrong, Pattern: regexp.MustCompile(`(?i)service auditor'?s report`)},
		{Name: "management_assertion", Weight: docextract.MarkerStrong, Pattern: regexp.MustCompile(`(?i)management'?s? assertion|assertion of management`)},
		{Name: "trust_services_criteria", Weight: docextract.MarkerStrong, Pattern: regexp.MustCompile(`(?i)trust services criteria|\bAICPA\b`)},
		{Name: "system_description", Weight: docextract.MarkerStrong, Pattern: regexp.MustCompile(`(?i)description of (the|its|[a-z]+'?s) system|system description`)},
		{Name: "criteria_ids", Weight: docextract.MarkerStrong, Pattern: regexp.MustCompile(`\b(CC|A|C|PI|P)\d{1,2}\.\d{1,2}\b`)},
		{Name: "type2_testing", Weight: docextract.MarkerStrong, Pattern: regexp.MustCompile(`(?i)tests? of controls|results of tests|\bno (exceptions|deviations) noted\b|\b(exceptions?|deviations?) noted\s*[.:;]|tests? of operating effectiveness`)},
		{Name: "reporting_period", Weight: docextract.MarkerSupporting, Pattern: regexp.MustCompile(`(?i)for the period|throughout the period|from [a-z]+ \d{1,2}, \d{4},? (through|to) [a-z]+ \d{1,2}, \d{4}`)},
		{Name: "as_of_date", Weight: docextract.MarkerSupporting, Pattern: regexp.MustCompile(`(?i)\bas of [a-z]+ \d{1,2}, \d{4}`)},
		{Name: "trust_categories", Weight: docextract.MarkerSupporting, Pattern: regexp.MustCompile(`(?i)\b(availability|confidentiality|processing integrity|privacy)\b`)},
		{Name: "auditor_cpa", Weight: docextract.MarkerSupporting, Pattern: regexp.MustCompile(`(?i)\bCPAs?\b|certified public accountants?`)},
		{Name: "cuecs", Weight: docextract.MarkerSupporting, Pattern: regexp.MustCompile(`(?i)complementary user entity controls|\bCUECs?\b|subservice organizations?`)},
		{Name: "type1_title", Weight: docextract.MarkerSupporting, Pattern: regexp.MustCompile(`(?i)\bSOC\W{0,3}2\W{0,3}Type\s*(1|I)\b`)},
		{Name: "type2_title", Weight: docextract.MarkerSupporting, Pattern: regexp.MustCompile(`(?i)\bSOC\W{0,3}2\W{0,3}Type\s*(2|II)\b`)},
		{Name: "design_only", Weight: docextract.MarkerSupporting, Pattern: regexp.MustCompile(`(?i)suitability of (the )?design of|not perform any procedures regarding the operating effectiveness`)},
		{Name: "opinion_stated", Weight: docextract.MarkerSupporting, Pattern: regexp.MustCompile(`(?i)\bin our opinion\b`)},
		{Name: "opinion_qualified", Weight: docextract.MarkerSupporting, Pattern: regexp.MustCompile(`(?i)\bqualified opinion\b|except for the (possible )?effects? of the matters?|except for the matters? (described|discussed|referred)`)},
		{Name: "opinion_adverse", Weight: docextract.MarkerSupporting, Pattern: regexp.MustCompile(`(?i)\badverse opinion\b`)},
		{Name: "opinion_disclaimer", Weight: docextract.MarkerSupporting, Pattern: regexp.MustCompile(`(?i)\bdisclaimer of opinion\b|do not express an opinion on (the|[a-z.]+'s) description|unable to obtain sufficient appropriate evidence to provide a basis for an opinion`)},
	},
}

// opinion reads the auditor's opinion from the report's headings and wording, most severe first; a
// report that states an opinion with none of the modified opinion language is unqualified
func opinion(found map[string]bool) docextract.Opinion {
	switch {
	case found["opinion_adverse"]:
		return OpinionAdverse
	case found["opinion_disclaimer"]:
		return OpinionDisclaimer
	case found["opinion_qualified"]:
		return OpinionQualified
	case found["opinion_stated"]:
		return OpinionUnqualified
	default:
		return docextract.OpinionUnknown
	}
}

// subtype calls a report Type 1 only on positive evidence with no sign of Type 2 testing, and falls
// back to Type 2 when unsure so the reviews and findings checks still apply
func subtype(found map[string]bool) docextract.ReportType {
	switch {
	case found["type2_title"] || found["type2_testing"]:
		return ReportTypeTwo
	case found["type1_title"]:
		return ReportTypeOne
	case found["design_only"] && found["as_of_date"] && !found["reporting_period"]:
		return ReportTypeOne
	default:
		return ReportTypeTwo
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
