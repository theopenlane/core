package oidcgeneric

import (
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/oauth"
	"github.com/theopenlane/core/internal/integrations/types"
)

// TypeOIDCGeneric identifies the generic OIDC provider
const TypeOIDCGeneric = types.ProviderType("oidc_generic")

// Builder returns the generic OIDC provider builder
func Builder() providers.Builder {
	return oauth.Builder(TypeOIDCGeneric)
}
