package enums

import "io"

// OAuthGrantKind is who an installed OAuth application acts as
type OAuthGrantKind string

var (
	// OAuthGrantKindUser is a personal connection that acts as the approving user, narrowed by its scopes
	OAuthGrantKindUser OAuthGrantKind = "USER"
	// OAuthGrantKindOrganization is an organization install that acts as a service subject, like an API token
	OAuthGrantKindOrganization OAuthGrantKind = "ORGANIZATION"
	// OAuthGrantKindInvalid represents an invalid grant kind
	OAuthGrantKindInvalid OAuthGrantKind = "INVALID"
)

var oauthGrantKindValues = []OAuthGrantKind{OAuthGrantKindUser, OAuthGrantKindOrganization}

// Values returns a slice of strings that represents all the possible values of the OAuthGrantKind enum
func (OAuthGrantKind) Values() []string { return stringValues(oauthGrantKindValues) }

// String returns the OAuthGrantKind as a string
func (r OAuthGrantKind) String() string { return string(r) }

// ToOAuthGrantKind returns the grant kind enum based on string input
func ToOAuthGrantKind(r string) *OAuthGrantKind {
	return parse(r, oauthGrantKindValues, &OAuthGrantKindInvalid)
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r OAuthGrantKind) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *OAuthGrantKind) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
