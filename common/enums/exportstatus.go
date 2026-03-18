package enums

import "io"

// ExportStatus is a custom type representing the various states of ExportStatus.
type ExportStatus string

var (
	// ExportStatusPending indicates the pending.
	ExportStatusPending ExportStatus = "PENDING"
	// ExportStatusFailed indicates the failed.
	ExportStatusFailed ExportStatus = "FAILED"
	// ExportStatusReady indicates the ready.
	ExportStatusReady ExportStatus = "READY"
	// ExportStatusNodata indicates the nodata.
	ExportStatusNodata ExportStatus = "NODATA"
	// ExportStatusInvalid is used when an unknown or unsupported value is provided.
	ExportStatusInvalid ExportStatus = "EXPORTSTATUS_INVALID"
)

var exportStatusValues = []ExportStatus{
	ExportStatusPending,
	ExportStatusFailed,
	ExportStatusReady,
	ExportStatusNodata,
}

// Values returns a slice of strings representing all valid ExportStatus values.
func (ExportStatus) Values() []string { return stringValues(exportStatusValues) }

// String returns the string representation of the ExportStatus value.
func (r ExportStatus) String() string { return string(r) }

// ToExportStatus converts a string to its corresponding ExportStatus enum value.
func ToExportStatus(r string) *ExportStatus {
	return parse(r, exportStatusValues, &ExportStatusInvalid)
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r ExportStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *ExportStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
