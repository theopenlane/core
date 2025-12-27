package enums

import (
	"fmt"
	"io"
	"strings"
)

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

// Values returns a slice of strings representing all valid ExportStatus values.
func (ExportStatus) Values() []string {
	return []string{
		string(ExportStatusPending),
		string(ExportStatusFailed),
		string(ExportStatusReady),
		string(ExportStatusNodata),
	}
}

// String returns the string representation of the ExportStatus value.
func (r ExportStatus) String() string {
	return string(r)
}

// ToExportStatus converts a string to its corresponding ExportStatus enum value.
func ToExportStatus(r string) *ExportStatus {
	switch strings.ToUpper(r) {
	case ExportStatusPending.String():
		return &ExportStatusPending
	case ExportStatusFailed.String():
		return &ExportStatusFailed
	case ExportStatusReady.String():
		return &ExportStatusReady
	case ExportStatusNodata.String():
		return &ExportStatusNodata
	default:
		return &ExportStatusInvalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r ExportStatus) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *ExportStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for ExportStatus, got: %T", v) //nolint:err113
	}

	*r = ExportStatus(str)

	return nil
}
