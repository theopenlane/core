package enums

import (
	"fmt"
	"io"
	"strings"
)

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

// Values returns a slice of strings that represents all the possible values of the AuthProvider enum.
// Possible default values are "CREDENTIALS", "GOOGLE", "GITHUB", "WEBAUTHN", and "OIDC"
func (AuthProvider) Values() (kinds []string) {
	for _, s := range []AuthProvider{AuthProviderCredentials, AuthProviderGoogle, AuthProviderGitHub, AuthProviderWebauthn, AuthProviderOIDC} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the AuthProvider as a string
func (r AuthProvider) String() string {
	return string(r)
}

// ToAuthProvider returns the AuthProvider based on string input
func ToAuthProvider(r string) *AuthProvider {
	switch r := strings.ToUpper(r); r {
	case AuthProviderCredentials.String():
		return &AuthProviderCredentials
	case AuthProviderGoogle.String():
		return &AuthProviderGoogle
	case AuthProviderGitHub.String():
		return &AuthProviderGitHub
	case AuthProviderWebauthn.String():
		return &AuthProviderWebauthn
	case AuthProviderOIDC.String():
		return &AuthProviderOIDC
	default:
		return &AuthProviderInvalid
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r AuthProvider) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *AuthProvider) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for AuthProvider, got: %T", v) //nolint:err113
	}

	*r = AuthProvider(str)

	return nil
}
