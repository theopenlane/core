package enums

import "io"

type SSOProvider string

var (
	SSOProviderOkta            SSOProvider = "OKTA"
	SSOProviderOneLogin        SSOProvider = "ONE_LOGIN"
	SSOProviderGoogleWorkspace SSOProvider = "GOOGLE_WORKSPACE"
	SSOProviderSlack           SSOProvider = "SLACK"
	SSOProviderGithub          SSOProvider = "GITHUB"
	SSOProviderEntraID         SSOProvider = "MICROSOFT_ENTRA_ID"
	SSOProviderGenericOIDC     SSOProvider = "GENERIC_OIDC"
	SSOProviderNone            SSOProvider = "NONE"
	SSOProviderInvalid         SSOProvider = "INVALID"
)

var ssoProviderValues = []SSOProvider{
	SSOProviderOkta,
	SSOProviderOneLogin,
	SSOProviderGoogleWorkspace,
	SSOProviderSlack,
	SSOProviderGithub,
	SSOProviderEntraID,
	SSOProviderGenericOIDC,
	SSOProviderNone,
}

// Values returns all possible SSOProvider values.
func (SSOProvider) Values() []string { return stringValues(ssoProviderValues) }

// String returns the SSOProvider as a string
func (r SSOProvider) String() string { return string(r) }

// ToSSOProvider returns the SSOProvider based on string input
func ToSSOProvider(r string) *SSOProvider { return parse(r, ssoProviderValues, &SSOProviderInvalid) }

// MarshalGQL implement the Marshaler interface for gqlgen
func (r SSOProvider) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *SSOProvider) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
