package oidcgeneric

import (
	"context"

	"github.com/theopenlane/core/common/integrations/helpers"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	// ClientOIDCAPI identifies the OIDC HTTP client used for userinfo calls.
	ClientOIDCAPI types.ClientName = "api"
)

// oidcClientDescriptors returns the client descriptors published by the generic OIDC provider.
func oidcClientDescriptors() []types.ClientDescriptor {
	return helpers.DefaultClientDescriptors(TypeOIDCGeneric, ClientOIDCAPI, "OIDC userinfo HTTP client", buildOIDCClient)
}

// buildOIDCClient constructs an authenticated OIDC client.
func buildOIDCClient(_ context.Context, payload types.CredentialPayload, _ map[string]any) (any, error) {
	token, err := helpers.OAuthTokenFromPayload(payload, string(TypeOIDCGeneric))
	if err != nil {
		return nil, err
	}

	return helpers.NewAuthenticatedClient(token, nil), nil
}
