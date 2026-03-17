package microsoftteams

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// teamsGraphBaseURL is the base URL for the Microsoft Graph API v1.0 endpoint used by Teams operations
const teamsGraphBaseURL = "https://graph.microsoft.com/v1.0/"

// Client builds Microsoft Graph API clients for one installation
type Client struct{}

// Build constructs the Microsoft Graph API client for one installation
func (Client) Build(_ context.Context, req types.ClientBuildRequest) (any, error) {
	if req.Credential.OAuthAccessToken == "" {
		return nil, ErrOAuthTokenMissing
	}

	return providerkit.NewAuthenticatedClient(teamsGraphBaseURL, req.Credential.OAuthAccessToken, nil), nil
}

// FromAny casts a registered client instance to the authenticated HTTP client type
func (Client) FromAny(value any) (*providerkit.AuthenticatedClient, error) {
	c, ok := value.(*providerkit.AuthenticatedClient)
	if !ok {
		return nil, ErrClientType
	}

	return c, nil
}
