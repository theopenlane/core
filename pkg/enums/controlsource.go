package enums

import (
	"fmt"
	"io"
	"strings"
)

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

// Values returns a slice of strings that represents all the possible values of the ControlSource enum.
// Possible default values are "FRAMEWORK", "TEMPLATE", "USER_DEFINED", and "IMPORTED".
func (ControlSource) Values() (kinds []string) {
	for _, s := range []ControlSource{ControlSourceFramework, ControlSourceTemplate, ControlSourceUserDefined, ControlSourceImport} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the ControlSource as a string
func (r ControlSource) String() string {
	return string(r)
}

// ToControlSource returns the control source enum based on string input
func ToControlSource(r string) *ControlSource {
	switch r := strings.ToUpper(r); r {
	case ControlSourceFramework.String():
		return &ControlSourceFramework
	case ControlSourceTemplate.String():
		return &ControlSourceTemplate
	case ControlSourceUserDefined.String():
		return &ControlSourceUserDefined
	case ControlSourceImport.String():
		return &ControlSourceImport
	default:
		return &ControlSourceInvalid
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r ControlSource) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *ControlSource) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for ControlSource, got: %T", v) //nolint:err113
	}

	*r = ControlSource(str)

	return nil
}
