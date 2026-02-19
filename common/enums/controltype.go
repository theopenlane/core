package enums

import "io"

// ControlType is a custom type representing the type of a control.
type ControlType string

var (
	// ControlTypePreventative is designed to prevent an event from occurring
	ControlTypePreventative ControlType = "PREVENTATIVE"
	// ControlTypeDetective is designed to detect an event that has occurred
	ControlTypeDetective ControlType = "DETECTIVE"
	// ControlTypeCorrective is designed to detect and correct an event that has occurred
	ControlTypeCorrective ControlType = "CORRECTIVE"
	// ControlTypeDeterrent acts as a deterrent to prevent an event from occurring
	ControlTypeDeterrent ControlType = "DETERRENT"
	// ControlTypeInvalid is used when the control type is invalid
	ControlTypeInvalid ControlType = "INVALID"
)

var controlTypeValues = []ControlType{
	ControlTypePreventative,
	ControlTypeDetective,
	ControlTypeCorrective,
	ControlTypeDeterrent,
}

// Values returns a slice of strings that represents all the possible values of the ControlType enum.
// Possible default values are "PREVENTATIVE", "DETECTIVE", "CORRECTIVE", and "DETERRENT".
func (ControlType) Values() []string { return stringValues(controlTypeValues) }

// String returns the ControlType as a string
func (r ControlType) String() string { return string(r) }

// ToControlType returns the control type enum based on string input
func ToControlType(r string) *ControlType {
	return parse(r, controlTypeValues, &ControlTypeInvalid)
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r ControlType) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *ControlType) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
