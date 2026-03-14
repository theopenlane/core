package azuresecuritycenter

import (
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrationsv2/definition"
	"github.com/theopenlane/core/internal/integrationsv2/types"
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

var (
	definitionSpec   = types.DefinitionSpec{
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
	}

	userInputSchema  = providerkit.SchemaFrom[userInput]()
	credentialSchema = providerkit.SchemaFrom[credential]()
)

// def implements definition.Assembler for the Azure Security Center integration
type def struct{}

// Builder returns the Azure Security Center definition builder
func Builder() definition.Builder {
	return definition.FromAssembler(&def{})
}

func (d *def) Spec() types.DefinitionSpec { return definitionSpec }

func (d *def) OperatorConfig() *types.OperatorConfigRegistration { return nil }

func (d *def) UserInput() *types.UserInputRegistration {
	return &types.UserInputRegistration{Schema: userInputSchema}
}

func (d *def) Credentials() *types.CredentialRegistration {
	return &types.CredentialRegistration{
		Schema:  credentialSchema,
		Persist: types.CredentialPersistModeKeystore,
	}
}

func (d *def) Auth() *types.AuthRegistration { return nil }

func (d *def) Clients() []types.ClientRegistration {
	return []types.ClientRegistration{
		{
			Name:        "api",
			Description: "Azure management API client for Defender for Cloud",
			Build:       buildAzureSecurityClient,
		},
	}
}

func (d *def) Operations() []types.OperationRegistration {
	return []types.OperationRegistration{
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
	}
}

func (d *def) Mappings() []types.MappingRegistration { return nil }
func (d *def) Webhooks() []types.WebhookRegistration { return nil }
