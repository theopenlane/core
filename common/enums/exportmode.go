package enums

import (
	"fmt"
	"io"
	"strings"
)

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

// Values returns a slice of strings representing all valid ExportMode values.
func (ExportMode) Values() []string {
	return []string{
		string(ExportModeFlat),
		string(ExportModeFolder),
	}
}

// String returns the string representation of the ExportMode value.
func (r ExportMode) String() string {
	return string(r)
}

// ToExportMode converts a string to its corresponding ExportMode enum value.
func ToExportMode(r string) *ExportMode {
	switch strings.ToUpper(r) {
	case ExportModeFlat.String():
		return &ExportModeFlat
	case ExportModeFolder.String():
		return &ExportModeFolder
	default:
		return &ExportModeInvalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r ExportMode) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *ExportMode) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for ExportMode, got: %T", v) //nolint:err113
	}

	*r = ExportMode(str)

	return nil
}
