package azuresecuritycenter

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
)

// userInput holds installation-specific configuration collected from the user
type userInput struct {
	Label         string `json:"label,omitempty"         jsonschema:"title=Installation Label"`
	ResourceGroup string `json:"resourceGroup,omitempty" jsonschema:"title=Resource Group"`
	WorkspaceID   string `json:"workspaceId,omitempty"   jsonschema:"title=Log Analytics Workspace ID"`
}

// credential holds the Azure Security Center client credentials for one installation
type credential struct {
	TenantID       string `json:"tenantId"                jsonschema:"required,title=Tenant ID"`
	ClientID       string `json:"clientId"                jsonschema:"required,title=Client ID"`
	ClientSecret   string `json:"clientSecret"            jsonschema:"required,title=Client Secret"`
	SubscriptionID string `json:"subscriptionId"          jsonschema:"required,title=Subscription ID"`
	ResourceGroup  string `json:"resourceGroup,omitempty" jsonschema:"title=Resource Group"`
	WorkspaceID    string `json:"workspaceId,omitempty"   jsonschema:"title=Log Analytics Workspace ID"`
}

// Builder returns the Azure Security Center definition builder
func Builder() definition.Builder {
	return definition.BuilderFunc(func(_ context.Context) (types.Definition, error) {
		return types.Definition{
			Spec: types.DefinitionSpec{
				ID:          "def_01K0AZSECC000000000000000001",
				Slug:        "azure_security_center",
				Version:     "v1",
				Family:      "azure",
				DisplayName: "Microsoft Defender for Cloud",
				Description: "Collect Microsoft Defender for Cloud pricing and plan metadata from an Azure subscription for security posture visibility.",
				Category:    "compliance",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/azure_security_center/overview",
				Labels:      map[string]string{"vendor": "microsoft", "product": "defender-for-cloud"},
				Active:      false,
				Visible:     true,
			},
			UserInput: &types.UserInputRegistration{
				Schema: providerkit.SchemaFrom[userInput](),
			},
			Credentials: &types.CredentialRegistration{
				Schema:  providerkit.SchemaFrom[credential](),
				Persist: types.CredentialPersistModeKeystore,
			},
			Clients: []types.ClientRegistration{
				{
					Name:        "api",
					Description: "Azure management API client for Defender for Cloud",
					Build:       buildAzureSecurityClient,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        "health.default",
					Kind:        types.OperationKindHealth,
					Description: "Call Azure Security Center pricings API to verify access",
					Topic:       gala.TopicName("integration.azure_security_center.health.default"),
					Client:      "api",
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      runHealthOperation,
				},
				{
					Name:        "security.pricing_overview",
					Kind:        types.OperationKindCollect,
					Description: "Collect plan and pricing metadata for Microsoft Defender for Cloud",
					Topic:       gala.TopicName("integration.azure_security_center.security.pricing_overview"),
					Client:      "api",
					Policy:      types.ExecutionPolicy{MaxRetries: 3, Idempotent: true},
					Handle:      runSecurityPricingOperation,
				},
			},
		}, nil
	})
}
