package azureentraid

import (
	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// Label is the user-defined display label for the installation
	Label string `json:"label,omitempty" jsonschema:"title=Installation Label"`
	// TenantID is the Azure Entra ID tenant identifier
	TenantID string `json:"tenantId,omitempty" jsonschema:"title=Tenant ID"`
}

// Builder returns the Azure Entra ID definition builder with the supplied operator config applied
func Builder(cfg Config) definition.Builder {
	return definition.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Slug:        Slug,
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
				Schema: providerkit.SchemaFrom[UserInput](),
			},
			Auth: &types.AuthRegistration{
				StartPath:    "/v1/integrations/oauth/start",
				CallbackPath: "/v1/integrations/oauth/callback",
				OAuth: &types.OAuthPublicConfig{
					ClientID:    cfg.ClientID,
					AuthURL:     azureAuthURL,
					TokenURL:    azureTokenURL,
					RedirectURI: cfg.RedirectURL,
					Scopes:      azureEntraScopes,
				},
				ClientSecret: cfg.ClientSecret,
			},
			Clients: []types.ClientRegistration{
				{
					Ref:         EntraClient.ID(),
					Description: "Microsoft Graph API client",
					Build:       Client{}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Call Microsoft Graph /organization to verify tenant access",
					Topic:       HealthDefaultOperation.Topic(Slug),
					ClientRef:   EntraClient.ID(),
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      HealthCheck{}.Handle(Client{}),
				},
				{
					Name:        DirectoryInspectOperation.Name(),
					Description: "Collect basic tenant metadata via Microsoft Graph",
					Topic:       DirectoryInspectOperation.Topic(Slug),
					ClientRef:   EntraClient.ID(),
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      DirectoryInspect{}.Handle(Client{}),
				},
			},
		}, nil
	})
}
