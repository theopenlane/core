package enums

import (
	"fmt"
	"io"
	"strings"
)

// ExportFormat is a custom type representing the various states of ExportFormat.
type ExportFormat string

var (
	// ExportFormatCsv indicates the csv.
	ExportFormatCsv ExportFormat = "CSV"
	// ExportFormatInvalid is used when an unknown or unsupported value is provided.
	ExportFormatInvalid ExportFormat = "EXPORTFORMAT_INVALID"
)

// Values returns a slice of strings representing all valid ExportFormat values.
func (ExportFormat) Values() []string {
	return []string{
		string(ExportFormatCsv),
	}
}

// String returns the string representation of the ExportFormat value.
func (r ExportFormat) String() string {
	return string(r)
}

// ToExportFormat converts a string to its corresponding ExportFormat enum value.
func ToExportFormat(r string) *ExportFormat {
	switch strings.ToUpper(r) {
	case ExportFormatCsv.String():
		return &ExportFormatCsv
	default:
		return &ExportFormatInvalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r ExportFormat) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *ExportFormat) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for ExportFormat, got: %T", v) //nolint:err113
	}

	*r = ExportFormat(str)

	return nil
}
