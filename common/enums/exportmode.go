package enums

import "io"

// ExportMode is a custom type representing the various states of ExportMode.
type ExportMode string

var (
	// ExportModeFlat indicates the flat.
	ExportModeFlat ExportMode = "FLAT"
	// ExportModeFolder indicates the folder.
	ExportModeFolder ExportMode = "FOLDER"
	// ExportModeInvalid is used when an unknown or unsupported value is provided.
	ExportModeInvalid ExportMode = "EXPORTMODE_INVALID"
)

var exportModeValues = []ExportMode{
	ExportModeFlat,
	ExportModeFolder,
}

// Values returns a slice of strings representing all valid ExportMode values.
func (ExportMode) Values() []string { return stringValues(exportModeValues) }

// String returns the string representation of the ExportMode value.
func (r ExportMode) String() string { return string(r) }

// ToExportMode converts a string to its corresponding ExportMode enum value.
func ToExportMode(r string) *ExportMode { return parse(r, exportModeValues, &ExportModeInvalid) }

// MarshalGQL implements the gqlgen Marshaler interface.
func (r ExportMode) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *ExportMode) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
