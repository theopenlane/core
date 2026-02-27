package azureentraid

import (
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/oauth"
)

// TypeAzureEntraID identifies the Azure Entra ID provider
const TypeAzureEntraID = types.ProviderType("azureentraid")

// Builder returns the Azure Entra ID provider builder
func Builder() providers.Builder {
	return oauth.Builder(TypeAzureEntraID, oauth.WithOperations(azureOperations()), oauth.WithClientDescriptors(azureEntraClientDescriptors()))
}
