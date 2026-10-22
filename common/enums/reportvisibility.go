package enums

import "io"

// ReportVisibility represents the visibility of a report.
type ReportVisibility string

var (
	// ReportVisibilityPrivate restricts access to the creator.
	ReportVisibilityPrivate      ReportVisibility = "PRIVATE"
	// ReportVisibilityRestricted limits access to report memberships and permitted groups.
	ReportVisibilityRestricted   ReportVisibility = "RESTRICTED"
	// ReportVisibilityOrganization allows organization members to view the report.
	ReportVisibilityOrganization ReportVisibility = "ORGANIZATION"
	// ReportVisibilityInvalid is used when an unknown or unsupported value is provided.
	ReportVisibilityInvalid      ReportVisibility = "REPORTVISIBILITY_INVALID"
)

var reportVisibilityValues = []ReportVisibility{
	ReportVisibilityPrivate,
	ReportVisibilityRestricted,
	ReportVisibilityOrganization,
}

// Values returns a slice of strings representing all valid ReportVisibility values.
func (ReportVisibility) Values() []string { return stringValues(reportVisibilityValues) }

// String returns the string representation of the ReportVisibility value.
func (r ReportVisibility) String() string { return string(r) }

// ToReportVisibility converts a string to its corresponding ReportVisibility enum value.
func ToReportVisibility(r string) *ReportVisibility {
	return parse(r, reportVisibilityValues, &ReportVisibilityInvalid)
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r ReportVisibility) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *ReportVisibility) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
