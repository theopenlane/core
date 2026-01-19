package enums

import (
	"fmt"
	"io"
	"strings"
)

// TrustCenterNDARequestAccessLevel is a custom type for NDA request access level
type TrustCenterNDARequestAccessLevel string

var (
	// TrustCenterNDARequestAccessLevelFull indicates full access to all documents
	TrustCenterNDARequestAccessLevelFull TrustCenterNDARequestAccessLevel = "FULL"
	// TrustCenterNDARequestAccessLevelLimited indicates limited access to specific documents
	TrustCenterNDARequestAccessLevelLimited TrustCenterNDARequestAccessLevel = "LIMITED"
	// TrustCenterNDARequestAccessLevelInvalid indicates the access level is invalid
	TrustCenterNDARequestAccessLevelInvalid TrustCenterNDARequestAccessLevel = "INVALID"
)

// Values returns a slice of strings that represents all the possible values of the TrustCenterNDARequestAccessLevel enum
func (TrustCenterNDARequestAccessLevel) Values() (kinds []string) {
	for _, s := range []TrustCenterNDARequestAccessLevel{
		TrustCenterNDARequestAccessLevelFull,
		TrustCenterNDARequestAccessLevelLimited,
	} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the access level as a string
func (r TrustCenterNDARequestAccessLevel) String() string {
	return string(r)
}

// ToTrustCenterNDARequestAccessLevel returns the access level enum based on string input
func ToTrustCenterNDARequestAccessLevel(r string) *TrustCenterNDARequestAccessLevel {
	switch r := strings.ToUpper(r); r {
	case TrustCenterNDARequestAccessLevelFull.String():
		return &TrustCenterNDARequestAccessLevelFull
	case TrustCenterNDARequestAccessLevelLimited.String():
		return &TrustCenterNDARequestAccessLevelLimited
	default:
		return nil
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r TrustCenterNDARequestAccessLevel) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *TrustCenterNDARequestAccessLevel) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for TrustCenterNDARequestAccessLevel, got: %T", v) //nolint:err113
	}

	*r = TrustCenterNDARequestAccessLevel(str)

	return nil
}
