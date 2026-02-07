package oidcgeneric

import (
	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	// ClientOIDCAPI identifies the OIDC HTTP client used for userinfo calls.
	ClientOIDCAPI types.ClientName = "api"
)

// oidcClientDescriptors returns the client descriptors published by the generic OIDC provider.
func oidcClientDescriptors() []types.ClientDescriptor {
	return auth.DefaultClientDescriptors(TypeOIDCGeneric, ClientOIDCAPI, "OIDC userinfo HTTP client", auth.OAuthClientBuilder(nil))
}
