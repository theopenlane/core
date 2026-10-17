package enums

import "io"

// ReportStatus is a custom type representing the various states of ReportStatus.
type ReportStatus string

var (
	// ReportStatusDraft indicates the draft.
	ReportStatusDraft ReportStatus = "DRAFT"
	// ReportStatusSaved indicates the saved.
	ReportStatusSaved ReportStatus = "SAVED"
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
func ToReportStatus(r string) *ReportStatus { return parse(r, reportStatusValues, &ReportStatusInvalid) }

// MarshalGQL implements the gqlgen Marshaler interface.
func (r ReportStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *ReportStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
