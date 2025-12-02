package azuresecuritycenter

import (
	"github.com/theopenlane/shared/integrations/providers"
	"github.com/theopenlane/shared/integrations/providers/oauth"
	"github.com/theopenlane/shared/integrations/types"
)

// TypeAzureSecurityCenter identifies the Azure Security Center provider
const TypeAzureSecurityCenter = types.ProviderType("azure_security_center")

// Builder returns the Azure Security Center provider builder
func Builder() providers.Builder {
	return oauth.Builder(TypeAzureSecurityCenter, oauth.WithOperations(azureSecurityOperations()))
}
