package soc2

import (
	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/docextract/schema"
	"github.com/theopenlane/core/v2/pkg/docextract"
)

// KindName identifies the SOC 2 report kind
const KindName = "soc2_report"

// iota on which parts of the output to include
const (
	IncludeVendors = 1 << iota
	IncludeAssets
	IncludeControls
	IncludeGroups
	IncludeDomains
	IncludeReviews
	IncludeFindings
	IncludePlatforms
	IncludeSystemDetails
	IncludeContacts
	IncludeProgram
	// IncludeAssessment is a vendor risk read of the report, not part of an import
	IncludeAssessment
)

// The section names each part emits, for callers that compare against a stored section name
const (
	// ControlsPart is the section whose ref codes scope the control-dependent parts
	ControlsPart = "controls"
	// ReviewsPart holds the test procedures and results, one per control
	ReviewsPart = "reviews"
	// FindingsPart holds the exceptions and deviations noted against controls
	FindingsPart = "findings"
	// AssessmentPart holds the summary and risk rating of a vendor's report
	AssessmentPart = "assessment"
)

// ControlScopedParts are the sections scoped by the ref codes of the controls section, so they
// cannot be extracted until the controls land
var ControlScopedParts = []string{ReviewsPart, FindingsPart}

// OptInParts are only extracted when a scan asks for them, never by default
var OptInParts = []string{AssessmentPart}

// VendorReviewParts are the sections a scan of a vendor's report extracts when it asks for none
var VendorReviewParts = []string{AssessmentPart, FindingsPart}

// Parts lists every individually extractable SOC 2 section, in the order they should be requested
var Parts = []int{
	// controls first, the control-scoped parts cannot start until it lands
	IncludeControls,
	IncludeProgram,
	IncludeDomains,
	IncludeVendors,
	IncludeAssets,
	IncludeGroups,
	IncludePlatforms,
	IncludeSystemDetails,
	IncludeContacts,
	IncludeFindings,
	IncludeReviews,
	IncludeAssessment,
}

// partNames maps each part to the section name its schema emits
var partNames = map[int]string{
	IncludeProgram:       "programs",
	IncludeDomains:       "domains",
	IncludeVendors:       "entities",
	IncludeAssets:        "assets",
	IncludeGroups:        "groups",
	IncludePlatforms:     "platforms",
	IncludeSystemDetails: "systemdetails",
	IncludeContacts:      "contacts",
	IncludeControls:      "controls",
	IncludeFindings:      "findings",
	IncludeReviews:       "reviews",
	IncludeAssessment:    AssessmentPart,
}

// sectionSchemas maps each part to the response schema the model must follow
var sectionSchemas = map[int]*docextract.Schema{
	IncludeVendors:       EntitySchema,
	IncludeAssets:        AssetSchema,
	IncludeControls:      ControlSchema,
	IncludeGroups:        GroupSchema,
	IncludeDomains:       DomainSchema,
	IncludeReviews:       ReviewSchema,
	IncludeFindings:      FindingSchema,
	IncludePlatforms:     PlatformSchema,
	IncludeSystemDetails: SystemDetailsSchema,
	IncludeContacts:      ContactSchema,
	IncludeProgram:       ProgramSchema,
	IncludeAssessment:    schema.AssessmentSchema,
}

// PartName returns the section name a part produces, empty for an unknown part
func PartName(part int) string {
	return partNames[part]
}

// PartByName returns the part that produces the named section
func PartByName(name string) (int, bool) {
	for part, partName := range partNames {
		if partName == name {
			return part, true
		}
	}

	return 0, false
}

// SectionFields lists the fields the named section's items carry, so an import keeps only what the model was asked for
func SectionFields(name string) []string {
	part, ok := PartByName(name)
	if !ok {
		return nil
	}

	list := sectionSchemas[part].Properties[name]
	if list == nil || list.Items == nil {
		return nil
	}

	return lo.Keys(list.Items.Properties)
}

// PartNames lists the section names for every part in request order
func PartNames() []string {
	names := make([]string, 0, len(Parts))
	for _, part := range Parts {
		names = append(names, partNames[part])
	}

	return names
}
