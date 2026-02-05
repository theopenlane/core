package azureentraid

import (
	"context"

	"github.com/theopenlane/core/common/integrations/helpers"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	// ClientAzureEntraAPI identifies the Microsoft Graph API client for Entra ID.
	ClientAzureEntraAPI types.ClientName = "api"
)

// azureEntraClientDescriptors returns the client descriptors published by Azure Entra ID.
func azureEntraClientDescriptors() []types.ClientDescriptor {
	return helpers.DefaultClientDescriptors(TypeAzureEntraID, ClientAzureEntraAPI, "Microsoft Graph API client", buildAzureEntraClient)
}

// buildAzureEntraClient constructs an authenticated Graph API client.
func buildAzureEntraClient(_ context.Context, payload types.CredentialPayload, _ map[string]any) (any, error) {
	token, err := helpers.OAuthTokenFromPayload(payload, string(TypeAzureEntraID))
	if err != nil {
		return nil, err
	}

	return helpers.NewAuthenticatedClient(token, nil), nil
}
