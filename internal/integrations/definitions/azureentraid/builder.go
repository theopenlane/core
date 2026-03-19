package azureentraid

import (
	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder returns the Azure EntraID definition builder with the supplied operator config applied
func Builder(cfg Config) definition.Builder {
	return definition.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Slug:        Slug,
				Family:      "azure",
				DisplayName: "Azure EntraID",
				Description: "Connect to Microsoft Graph to validate tenant access and inspect Azure Entra ID organization metadata.",
				Category:    "identity",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/azure_entra_id/overview",
				Labels:      map[string]string{"vendor": "microsoft", "product": "EntraID"},
				Active:      false,
				Visible:     true,
			},
			OperatorConfig: &types.OperatorConfigRegistration{
				Schema: providerkit.SchemaFrom[Config](),
			},
			UserInput: &types.UserInputRegistration{
				Schema: providerkit.SchemaFrom[UserInput](),
			},
			Credentials: &types.CredentialRegistration{
				Schema: providerkit.SchemaFrom[CredentialSchema](),
			},
			Clients: []types.ClientRegistration{
				{
					Ref:         EntraCredential.ID(),
					Description: "Azure client credentials token credential for auth verification",
					Build:       CredentialClient{cfg: cfg}.Build,
				},
				{
					Ref:         EntraClient.ID(),
					Description: "Microsoft Graph service client for directory operations",
					Build:       GraphClient{cfg: cfg}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Verify Azure client credentials can acquire a token against Microsoft Graph",
					Topic:       HealthDefaultOperation.Topic(Slug),
					ClientRef:   EntraCredential.ID(),
					Handle:      HealthCheck{}.Handle(),
				},
				{
					Name:        DirectoryInspectOperation.Name(),
					Description: "Collect basic tenant metadata via Microsoft Graph",
					Topic:       DirectoryInspectOperation.Topic(Slug),
					ClientRef:   EntraClient.ID(),
					Handle:      DirectoryInspect{}.Handle(),
				},
			},
		}, nil
	})
}
