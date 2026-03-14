package azureentraid

import (
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrationsv2/definition"
	"github.com/theopenlane/core/internal/integrationsv2/types"
	"github.com/theopenlane/core/pkg/gala"
)

// userInput holds installation-specific configuration collected from the user
type userInput struct {
	Label    string `json:"label,omitempty"    jsonschema:"title=Installation Label"`
	TenantID string `json:"tenantId,omitempty" jsonschema:"title=Tenant ID"`
}

var (
	definitionSpec  = types.DefinitionSpec{
		ID:          "def_01K0AZENTRAID000000000000001",
		Slug:        "azure_entra_id",
		Version:     "v1",
		Family:      "azure",
		DisplayName: "Azure Entra ID",
		Description: "Connect to Microsoft Graph to validate tenant access and inspect Azure Entra ID organization metadata.",
		Category:    "identity",
		DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/azure_entra_id/overview",
		Labels:      map[string]string{"vendor": "microsoft", "product": "entra-id"},
		Active:      false,
		Visible:     true,
	}

	configSchema    = providerkit.SchemaFrom[Config]()
	userInputSchema = providerkit.SchemaFrom[userInput]()
)

// def implements definition.Assembler for the Azure Entra ID integration
type def struct {
	cfg Config
}

// Builder returns the Azure Entra ID definition builder with the supplied operator config applied
func Builder(cfg Config) definition.Builder {
	return definition.FromAssembler(&def{cfg: cfg})
}

func (d *def) Spec() types.DefinitionSpec { return definitionSpec }

func (d *def) OperatorConfig() *types.OperatorConfigRegistration {
	return &types.OperatorConfigRegistration{Schema: configSchema}
}

func (d *def) UserInput() *types.UserInputRegistration {
	return &types.UserInputRegistration{Schema: userInputSchema}
}

func (d *def) Credentials() *types.CredentialRegistration { return nil }

func (d *def) Auth() *types.AuthRegistration {
	return &types.AuthRegistration{
		Start:    startInstallAuth,
		Complete: completeInstallAuth,
	}
}

func (d *def) Clients() []types.ClientRegistration {
	return []types.ClientRegistration{
		{
			Name:        "api",
			Description: "Microsoft Graph API client",
			Build:       buildGraphClient,
		},
	}
}

func (d *def) Operations() []types.OperationRegistration {
	return []types.OperationRegistration{
		{
			Name:        "health.default",
			Kind:        types.OperationKindHealth,
			Description: "Call Microsoft Graph /organization to verify tenant access",
			Topic:       gala.TopicName("integration.azure_entra_id.health.default"),
			Client:      "api",
			Policy:      types.ExecutionPolicy{Idempotent: true},
			Handle:      runHealthOperation,
		},
		{
			Name:        "directory.inspect",
			Kind:        types.OperationKindCollect,
			Description: "Collect basic tenant metadata via Microsoft Graph",
			Topic:       gala.TopicName("integration.azure_entra_id.directory.inspect"),
			Client:      "api",
			Policy:      types.ExecutionPolicy{Idempotent: true},
			Handle:      runDirectoryInspectOperation,
		},
	}
}

func (d *def) Mappings() []types.MappingRegistration { return nil }
func (d *def) Webhooks() []types.WebhookRegistration { return nil }
