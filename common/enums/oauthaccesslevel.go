package enums

import "io"

// OAuthAccessLevel is what a personal OAuth connection may do on behalf of its user
type OAuthAccessLevel string

var (
	// OAuthAccessLevelRead allows only reads, every mutation is rejected
	OAuthAccessLevelRead OAuthAccessLevel = "READ"
	// OAuthAccessLevelWrite allows reads and writes within the granted scopes
	OAuthAccessLevelWrite OAuthAccessLevel = "WRITE"
	// OAuthAccessLevelInvalid represents an invalid access level
	OAuthAccessLevelInvalid OAuthAccessLevel = "INVALID"
)

var oauthAccessLevelValues = []OAuthAccessLevel{OAuthAccessLevelRead, OAuthAccessLevelWrite}

// Values returns a slice of strings that represents all the possible values of the OAuthAccessLevel enum
func (OAuthAccessLevel) Values() []string { return stringValues(oauthAccessLevelValues) }

// String returns the OAuthAccessLevel as a string
func (r OAuthAccessLevel) String() string { return string(r) }

// ToOAuthAccessLevel returns the access level enum based on string input
func ToOAuthAccessLevel(r string) *OAuthAccessLevel {
	return parse(r, oauthAccessLevelValues, &OAuthAccessLevelInvalid)
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r OAuthAccessLevel) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *OAuthAccessLevel) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
