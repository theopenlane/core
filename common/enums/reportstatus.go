package enums

import "io"

// ReportStatus represents the status of a report.
type ReportStatus string

var (
	// ReportStatusDraft indicates a draft report.
	ReportStatusDraft   ReportStatus = "DRAFT"
	// ReportStatusSaved indicates a saved report.
	ReportStatusSaved   ReportStatus = "SAVED"
	// ReportStatusInvalid is used when an unknown or unsupported value is provided.
	ReportStatusInvalid ReportStatus = "REPORTSTATUS_INVALID"
)

var reportStatusValues = []ReportStatus{
	ReportStatusDraft,
	ReportStatusSaved,
}

// Values returns a slice of strings representing all valid ReportStatus values.
func (ReportStatus) Values() []string { return stringValues(reportStatusValues) }

// String returns the string representation of the ReportStatus value.
func (r ReportStatus) String() string { return string(r) }

// ToReportStatus converts a string to its corresponding ReportStatus enum value.
func ToReportStatus(r string) *ReportStatus {
	return parse(r, reportStatusValues, &ReportStatusInvalid)
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r ReportStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *ReportStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
