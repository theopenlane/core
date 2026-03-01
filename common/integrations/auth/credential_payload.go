package auth

import (
	"golang.org/x/oauth2"

	"github.com/zitadel/oidc/v3/pkg/oidc"

	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/common/models"
)

// BuildCredentialPayload constructs a credential payload from variadic credential options.
func BuildCredentialPayload(provider types.ProviderType, opts ...types.CredentialOption) (types.CredentialPayload, error) {
	return types.NewCredentialBuilder(provider).With(opts...).Build()
}

// BuildOAuthCredentialPayload builds a normalized OAuth credential payload.
func BuildOAuthCredentialPayload(provider types.ProviderType, token *oauth2.Token, claims *oidc.IDTokenClaims, opts ...types.CredentialOption) (types.CredentialPayload, error) {
	options := []types.CredentialOption{
		types.WithCredentialSet(models.CredentialSet{}),
		types.WithOAuthToken(token),
		types.WithCredentialKind(types.CredentialKindOAuthToken),
	}
	if claims != nil {
		options = append(options, types.WithOIDCClaims(claims))
	}
	options = append(options, opts...)

	return BuildCredentialPayload(provider, options...)
}

// BuildAPITokenCredentialPayload builds a normalized API token credential payload.
func BuildAPITokenCredentialPayload(provider types.ProviderType, token string, providerData map[string]any, opts ...types.CredentialOption) (types.CredentialPayload, error) {
	options := []types.CredentialOption{
		types.WithCredentialKind(types.CredentialKindAPIKey),
		types.WithCredentialSet(models.CredentialSet{
			APIToken:     token,
			ProviderData: CloneMetadata(providerData),
		}),
	}
	options = append(options, opts...)

	return BuildCredentialPayload(provider, options...)
}
