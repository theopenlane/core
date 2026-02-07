package azureentraid

import (
	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	// ClientAzureEntraAPI identifies the Microsoft Graph API client for Entra ID.
	ClientAzureEntraAPI types.ClientName = "api"
)

// azureEntraClientDescriptors returns the client descriptors published by Azure Entra ID.
func azureEntraClientDescriptors() []types.ClientDescriptor {
	return auth.DefaultClientDescriptors(TypeAzureEntraID, ClientAzureEntraAPI, "Microsoft Graph API client", auth.OAuthClientBuilder(nil))
}
