package oidcgeneric

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/config"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/oauth"
	"github.com/theopenlane/core/internal/integrations/types"
)

// TypeOIDCGeneric identifies the generic OIDC provider
const TypeOIDCGeneric = types.ProviderType("oidc_generic")

// Builder returns the generic OIDC provider builder
func Builder() providers.Builder {
	return providers.BuilderFunc{
		ProviderType: TypeOIDCGeneric,
		BuildFunc: func(ctx context.Context, spec config.ProviderSpec) (providers.Provider, error) {
			ops := oidcOperations(userInfoURL(spec))
			return oauth.New(ctx, spec, oauth.WithOperations(ops))
		},
	}
}

func userInfoURL(spec config.ProviderSpec) string {
	if spec.UserInfo != nil {
		return spec.UserInfo.URL
	}
	return ""
}
