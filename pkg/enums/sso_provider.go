package enums

import (
	"fmt"
	"io"
	"strings"
)

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

// Values returns all possible SSOProvider values.
func (SSOProvider) Values() (kinds []string) {
	for _, s := range []SSOProvider{SSOProviderOkta, SSOProviderOneLogin, SSOProviderGoogleWorkspace, SSOProviderSlack, SSOProviderGithub, SSOProviderNone} {
		kinds = append(kinds, string(s))
	}

	return
}

func (p SSOProvider) String() string { return string(p) }

func ToSSOProvider(in string) *SSOProvider {
	switch strings.ToUpper(in) {
	case SSOProviderOkta.String():
		return &SSOProviderOkta
	case SSOProviderOneLogin.String():
		return &SSOProviderOneLogin
	case SSOProviderGoogleWorkspace.String():
		return &SSOProviderGoogleWorkspace
	case SSOProviderSlack.String():
		return &SSOProviderSlack
	case SSOProviderGithub.String():
		return &SSOProviderGithub
	case SSOProviderNone.String():
		return &SSOProviderNone
	default:
		return &SSOProviderInvalid
	}
}

func (p SSOProvider) MarshalGQL(w io.Writer) { _, _ = w.Write([]byte(`"` + p.String() + `"`)) }

func (p *SSOProvider) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for SSOProvider, got: %T", v) //nolint:err113
	}

	*p = SSOProvider(str)

	return nil
}
