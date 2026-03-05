package azureentraid

import (
	"github.com/theopenlane/core/internal/integrations/auth"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	// ClientAzureEntraAPI identifies the Microsoft Graph API client for Entra ID.
	ClientAzureEntraAPI types.ClientName = "api"
)

// azureEntraClientDescriptors returns the client descriptors published by Azure Entra ID.
func azureEntraClientDescriptors() []types.ClientDescriptor {
	return providerkit.DefaultClientDescriptors(TypeAzureEntraID, ClientAzureEntraAPI, "Microsoft Graph API client", providerkit.TokenClientBuilder(auth.OAuthTokenFromPayload, nil))
}
