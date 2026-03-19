package azureentraid

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// azureEntraGraphBaseURL is the base URL for the Microsoft Graph API v1.0 endpoint
const azureEntraGraphBaseURL = "https://graph.microsoft.com/v1.0/"

// Client builds Microsoft Graph API clients for one installation
type Client struct{}

// Build constructs the Microsoft Graph API client for one installation
func (Client) Build(_ context.Context, req types.ClientBuildRequest) (any, error) {
	if req.Credential.OAuthAccessToken == "" {
		return nil, ErrOAuthTokenMissing
	}

	return providerkit.NewAuthenticatedClient(azureEntraGraphBaseURL, req.Credential.OAuthAccessToken, nil), nil
}

