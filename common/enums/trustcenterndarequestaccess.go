package enums

import "io"

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

var trustCenterNDARequestAccessLevelValues = []TrustCenterNDARequestAccessLevel{
	TrustCenterNDARequestAccessLevelFull,
	TrustCenterNDARequestAccessLevelLimited,
}

// Values returns a slice of strings that represents all the possible values of the TrustCenterNDARequestAccessLevel enum
func (TrustCenterNDARequestAccessLevel) Values() []string {
	return stringValues(trustCenterNDARequestAccessLevelValues)
}

// String returns the access level as a string
func (r TrustCenterNDARequestAccessLevel) String() string { return string(r) }

// ToTrustCenterNDARequestAccessLevel returns the access level enum based on string input
func ToTrustCenterNDARequestAccessLevel(r string) *TrustCenterNDARequestAccessLevel {
	return parse(r, trustCenterNDARequestAccessLevelValues, nil)
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r TrustCenterNDARequestAccessLevel) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *TrustCenterNDARequestAccessLevel) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
