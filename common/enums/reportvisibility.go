package enums

import "io"

type ReportVisibility string

var (
	ReportVisibilityPrivate      ReportVisibility = "PRIVATE"
	ReportVisibilityRestricted   ReportVisibility = "RESTRICTED"
	ReportVisibilityOrganization ReportVisibility = "ORGANIZATION"
	ReportVisibilityInvalid      ReportVisibility = "REPORTVISIBILITY_INVALID"
)

var reportVisibilityValues = []ReportVisibility{
	ReportVisibilityPrivate,
	ReportVisibilityRestricted,
	ReportVisibilityOrganization,
}

func (ReportVisibility) Values() []string { return stringValues(reportVisibilityValues) }

func (r ReportVisibility) String() string { return string(r) }

func ToReportVisibility(r string) *ReportVisibility {
	return parse(r, reportVisibilityValues, &ReportVisibilityInvalid)
}

func (r ReportVisibility) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

func (r *ReportVisibility) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
