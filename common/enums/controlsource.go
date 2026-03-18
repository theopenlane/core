package enums

import "io"

// ControlSource is a custom type representing the source of a control.
type ControlSource string

var (
	// ControlSourceFramework is used when the control comes from an official framework (e.g. NIST, ISO, etc.)
	ControlSourceFramework ControlSource = "FRAMEWORK"
	// ControlSourceTemplate is used when the control comes from a template
	ControlSourceTemplate ControlSource = "TEMPLATE"
	// ControlSourceUserDefined is used when the control is manually created by a user
	ControlSourceUserDefined ControlSource = "USER_DEFINED"
	// ControlSourceImport is used when the control is imported from another system
	ControlSourceImport ControlSource = "IMPORTED"
	// ControlSourceInvalid is used when the control source is invalid
	ControlSourceInvalid ControlSource = "INVALID"
)

var controlSourceValues = []ControlSource{
	ControlSourceFramework,
	ControlSourceTemplate,
	ControlSourceUserDefined,
	ControlSourceImport,
}

// Values returns a slice of strings that represents all the possible values of the ControlSource enum.
// Possible default values are "FRAMEWORK", "TEMPLATE", "USER_DEFINED", and "IMPORTED".
func (ControlSource) Values() []string { return stringValues(controlSourceValues) }

// String returns the ControlSource as a string
func (r ControlSource) String() string { return string(r) }

// ToControlSource returns the control source enum based on string input
func ToControlSource(r string) *ControlSource {
	return parse(r, controlSourceValues, &ControlSourceInvalid)
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r ControlSource) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *ControlSource) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
