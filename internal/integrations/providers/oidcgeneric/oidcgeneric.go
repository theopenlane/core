package oidcgeneric

import (
	"context"

	"github.com/theopenlane/core/common/integrations/config"
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/oauth"
)

// TypeOIDCGeneric identifies the generic OIDC provider
const TypeOIDCGeneric = types.ProviderType("oidcgeneric")

// Builder returns the generic OIDC provider builder
func Builder() providers.Builder {
	return providers.BuilderFunc{
		ProviderType: TypeOIDCGeneric,
		BuildFunc: func(_ context.Context, spec config.ProviderSpec) (providers.Provider, error) {
			ops := oidcOperations(userInfoURL(spec))
			return oauth.New(spec, oauth.WithOperations(ops), oauth.WithClientDescriptors(oidcClientDescriptors()))
		},
	}
}

// userInfoURL returns the configured userinfo endpoint when present
func userInfoURL(spec config.ProviderSpec) string {
	if spec.UserInfo != nil {
		return spec.UserInfo.URL
	}

	return ""
}
