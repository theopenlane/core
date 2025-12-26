package azureentraid

import (
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/pkg/integrations/providers"
	"github.com/theopenlane/core/pkg/integrations/providers/oauth"
)

// TypeAzureEntraID identifies the Azure Entra ID provider
const TypeAzureEntraID = types.ProviderType("azure_entra_id")

// Builder returns the Azure Entra ID provider builder
func Builder() providers.Builder {
	return oauth.Builder(TypeAzureEntraID, oauth.WithOperations(azureOperations()))
}
