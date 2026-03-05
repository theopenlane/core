package oidcgeneric

import (
	"github.com/theopenlane/core/internal/integrations/auth"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	// ClientOIDCAPI identifies the OIDC HTTP client used for userinfo calls.
	ClientOIDCAPI types.ClientName = "api"
)

// oidcClientDescriptors returns the client descriptors published by the generic OIDC provider.
func oidcClientDescriptors() []types.ClientDescriptor {
	return providerkit.DefaultClientDescriptors(TypeOIDCGeneric, ClientOIDCAPI, "OIDC userinfo HTTP client", providerkit.TokenClientBuilder(auth.OAuthTokenFromPayload, nil))
}
