package azureentraid

import (
	"github.com/theopenlane/shared/integrations/providers"
	"github.com/theopenlane/shared/integrations/providers/oauth"
	"github.com/theopenlane/shared/integrations/types"
)

// TypeAzureEntraID identifies the Azure Entra ID provider
const TypeAzureEntraID = types.ProviderType("azure_entra_id")

// Builder returns the Azure Entra ID provider builder
func Builder() providers.Builder {
	return oauth.Builder(TypeAzureEntraID, oauth.WithOperations(azureOperations()))
}
