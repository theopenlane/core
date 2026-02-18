package enums

import "io"

// ExportFormat is a custom type representing the various states of ExportFormat.
type ExportFormat string

var (
	// ExportFormatCsv indicates the csv.
	ExportFormatCsv ExportFormat = "CSV"
	// ExportFormatInvalid is used when an unknown or unsupported value is provided.
	ExportFormatInvalid ExportFormat = "EXPORTFORMAT_INVALID"
)

var exportFormatValues = []ExportFormat{ExportFormatCsv}

// Values returns a slice of strings representing all valid ExportFormat values.
func (ExportFormat) Values() []string { return stringValues(exportFormatValues) }

// String returns the string representation of the ExportFormat value.
func (r ExportFormat) String() string { return string(r) }

// ToExportFormat converts a string to its corresponding ExportFormat enum value.
func ToExportFormat(r string) *ExportFormat {
	return parse(r, exportFormatValues, &ExportFormatInvalid)
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r ExportFormat) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *ExportFormat) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
