package enums

import (
	"fmt"
	"io"
	"strings"
)

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

// Values returns a slice of strings that represents all the possible values of the ControlType enum.
// Possible default values are "PREVENTATIVE", "DETECTIVE", "CORRECTIVE", and "DETERRENT".
func (ControlType) Values() (kinds []string) {
	for _, s := range []ControlType{ControlTypePreventative, ControlTypeDetective, ControlTypeCorrective, ControlTypeDeterrent} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the ControlType as a string
func (r ControlType) String() string {
	return string(r)
}

// ToControlType returns the control type enum based on string input
func ToControlType(r string) *ControlType {
	switch r := strings.ToUpper(r); r {
	case ControlTypePreventative.String():
		return &ControlTypePreventative
	case ControlTypeDetective.String():
		return &ControlTypeDetective
	case ControlTypeCorrective.String():
		return &ControlTypeCorrective
	case ControlTypeDeterrent.String():
		return &ControlTypeDeterrent
	default:
		return &ControlTypeInvalid
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r ControlType) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *ControlType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for ControlType, got: %T", v) //nolint:err113
	}

	*r = ControlType(str)

	return nil
}
