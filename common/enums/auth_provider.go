package enums

import "io"

// AuthProvider is a custom type representing authentication providers.
type AuthProvider string

var (
	// Credentials provider is when the user authenticates with a username and password
	AuthProviderCredentials AuthProvider = "CREDENTIALS"
	// Google oauth2 provider for authentication
	AuthProviderGoogle AuthProvider = "GOOGLE"
	// Github oauth2 provider for authentication
	AuthProviderGitHub AuthProvider = "GITHUB"
	// Webauthn passkey provider for authentication
	AuthProviderWebauthn AuthProvider = "WEBAUTHN"
	// OIDC provider for authentication
	AuthProviderOIDC AuthProvider = "OIDC"
	// AuthProviderInvalid is the default value for the AuthProvider enum
	AuthProviderInvalid AuthProvider = "INVALID"
)

var authProviderValues = []AuthProvider{
	AuthProviderCredentials,
	AuthProviderGoogle,
	AuthProviderGitHub,
	AuthProviderWebauthn,
	AuthProviderOIDC,
}

// Values returns a slice of strings that represents all the possible values of the AuthProvider enum.
// Possible default values are "CREDENTIALS", "GOOGLE", "GITHUB", "WEBAUTHN", and "OIDC"
func (AuthProvider) Values() []string { return stringValues(authProviderValues) }

// String returns the AuthProvider as a string
func (r AuthProvider) String() string { return string(r) }

// ToAuthProvider returns the AuthProvider based on string input
func ToAuthProvider(r string) *AuthProvider {
	return parse(r, authProviderValues, &AuthProviderInvalid)
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r AuthProvider) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *AuthProvider) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
