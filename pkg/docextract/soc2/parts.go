package soc2

import (
	"google.golang.org/genai"

	"github.com/theopenlane/core/v2/pkg/docextract/schema"
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
	IncludeAll = IncludeVendors | IncludeAssets | IncludeControls | IncludeGroups | IncludeDomains | IncludeReviews | IncludeFindings | IncludePlatforms | IncludeSystemDetails | IncludeContacts | IncludeProgram
)

// The section names each part emits, for callers that compare against a stored section name
const (
	// ControlsPart is the section whose ref codes scope the control-dependent parts
	ControlsPart = "controls"
	// ReviewsPart holds the test procedures and results, one per control
	ReviewsPart = "reviews"
	// FindingsPart holds the exceptions and deviations noted against controls
	FindingsPart = "findings"
)

// ControlScopedParts are the sections scoped by the ref codes of the controls section, so they
// cannot be extracted until the controls land
var ControlScopedParts = []string{ReviewsPart, FindingsPart}

// Parts lists every individually extractable SOC 2 section, in the order they should be requested
var Parts = []int{
	IncludeProgram,
	IncludeDomains,
	IncludeVendors,
	IncludeAssets,
	IncludeGroups,
	IncludePlatforms,
	IncludeSystemDetails,
	IncludeContacts,
	IncludeControls,
	IncludeFindings,
	IncludeReviews,
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
}

// sectionSchemas maps each part to the response schema the model must follow
var sectionSchemas = map[int]*genai.Schema{
	IncludeAll:           schema.FullImportSchema,
	IncludeVendors:       schema.EntitySchema,
	IncludeAssets:        schema.AssetSchema,
	IncludeControls:      schema.ControlSchema,
	IncludeGroups:        schema.GroupSchema,
	IncludeDomains:       schema.DomainSchema,
	IncludeReviews:       schema.ReviewSchema,
	IncludeFindings:      schema.FindingSchema,
	IncludePlatforms:     schema.PlatformSchema,
	IncludeSystemDetails: schema.SystemDetailsSchema,
	IncludeContacts:      schema.ContactSchema,
	IncludeProgram:       schema.ProgramSchema,
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

// PartNames lists the section names for every part in request order
func PartNames() []string {
	names := make([]string, 0, len(Parts))
	for _, part := range Parts {
		names = append(names, partNames[part])
	}

	return names
}
