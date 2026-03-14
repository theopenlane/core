package azureentraid

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
)

// userInput holds installation-specific configuration collected from the user
type userInput struct {
	Label    string `json:"label,omitempty"    jsonschema:"title=Installation Label"`
	TenantID string `json:"tenantId,omitempty" jsonschema:"title=Tenant ID"`
}

// def holds operator config for the Azure Entra ID integration
type def struct {
	cfg Config
}

// Builder returns the Azure Entra ID definition builder with the supplied operator config applied
func Builder(cfg Config) definition.Builder {
	d := &def{cfg: cfg}
	return definition.BuilderFunc(func(_ context.Context) (types.Definition, error) {
		return types.Definition{
			Spec: types.DefinitionSpec{
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
			},
			OperatorConfig: &types.OperatorConfigRegistration{
				Schema: providerkit.SchemaFrom[Config](),
			},
			UserInput: &types.UserInputRegistration{
				Schema: providerkit.SchemaFrom[userInput](),
			},
			Auth: &types.AuthRegistration{
				Start:    d.startInstallAuth,
				Complete: d.completeInstallAuth,
			},
			Clients: []types.ClientRegistration{
				{
					Name:        "api",
					Description: "Microsoft Graph API client",
					Build:       buildGraphClient,
				},
			},
			Operations: []types.OperationRegistration{
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
			},
		}, nil
	})
}
