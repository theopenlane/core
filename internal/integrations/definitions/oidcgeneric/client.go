package oidcgeneric

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Client builds OIDC userinfo HTTP clients for one installation
type Client struct{}

// Build constructs the OIDC userinfo HTTP client for one installation
func (Client) Build(_ context.Context, req types.ClientBuildRequest) (any, error) {
	if req.Credential.OAuthAccessToken == "" {
		return nil, ErrOAuthTokenMissing
	}

	return providerkit.NewAuthenticatedClient("", req.Credential.OAuthAccessToken, nil), nil
}

