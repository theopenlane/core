package soc2

import (
	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/docextract/schema/custom"
	"github.com/theopenlane/core/v2/pkg/docextract"
)

// KindName identifies the SOC 2 report kind
const KindName = "soc2_report"

// The section names each part emits, for callers that compare against a stored section name
const (
	// ControlsPart is the section whose ref codes scope the control-dependent parts
	ControlsPart = "controls"
	// ProgramsPart holds the audited program the report covers
	ProgramsPart = "programs"
	// EntitiesPart holds the vendors and subservice organizations the report names
	EntitiesPart = "entities"
	// AssetsPart holds the systems and infrastructure in the audited environment
	AssetsPart = "assets"
	// GroupsPart holds the teams and roles responsible for the controls
	GroupsPart = "groups"
	// PlatformsPart holds the platforms the audited system runs on
	PlatformsPart = "platforms"
	// SystemDetailsPart holds the systems described in the report
	SystemDetailsPart = "systemdetails"
	// ContactsPart holds the people the report names
	ContactsPart = "contacts"
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

// Parts lists every individually extractable SOC 2 section, in the order they should be requested
var Parts = []string{
	// controls first, the control-scoped parts cannot start until it lands
	ControlsPart,
	ProgramsPart,
	EntitiesPart,
	AssetsPart,
	GroupsPart,
	PlatformsPart,
	SystemDetailsPart,
	ContactsPart,
	FindingsPart,
	ReviewsPart,
	AssessmentPart,
}

// sectionSchemas maps each part to the response schema the model must follow
var sectionSchemas = map[string]*docextract.Schema{
	EntitiesPart:      EntitySchema,
	AssetsPart:        AssetSchema,
	ControlsPart:      ControlSchema,
	GroupsPart:        GroupSchema,
	ReviewsPart:       ReviewSchema,
	FindingsPart:      FindingSchema,
	PlatformsPart:     PlatformSchema,
	SystemDetailsPart: SystemDetailsSchema,
	ContactsPart:      ContactSchema,
	ProgramsPart:      ProgramSchema,
	AssessmentPart:    custom.AssessmentSchema,
}

// Profile is what a scan extracts when it names no sections, decided by why the report is parsed
type Profile struct {
	// Parts are the sections extracted
	Parts []string
	// RequireAll fails the scan when any of its parts has no prompt configured, rather than skipping it
	RequireAll bool
}

// ImportProfile extracts the records an organization imports from its own report
var ImportProfile = Profile{Parts: lo.Without(Parts, AssessmentPart)}

// VendorReviewProfile extracts the risk read and exceptions of a vendor's report
var VendorReviewProfile = Profile{Parts: []string{AssessmentPart, FindingsPart}, RequireAll: true}

// IsPart reports whether name is a SOC 2 section
func IsPart(name string) bool {
	_, ok := sectionSchemas[name]

	return ok
}

// SectionFields lists the fields the named section's items carry, so an import keeps only what the model was asked for
func SectionFields(name string) []string {
	section, ok := sectionSchemas[name]
	if !ok {
		return nil
	}

	property := section.Properties[name]
	if property == nil {
		return nil
	}

	if property.Items != nil {
		property = property.Items
	}

	return lo.Keys(property.Properties)
}
