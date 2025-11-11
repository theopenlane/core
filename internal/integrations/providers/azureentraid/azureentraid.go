package azureentraid

import (
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/oauth"
	"github.com/theopenlane/core/internal/integrations/types"
)

// TypeAzureEntraID identifies the Azure Entra ID provider
const TypeAzureEntraID = types.ProviderType("azure_entra_id")

// Builder returns the Azure Entra ID provider builder
func Builder() providers.Builder {
	return oauth.Builder(TypeAzureEntraID)
}
