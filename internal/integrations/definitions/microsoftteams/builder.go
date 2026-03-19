package microsoftteams

import (
	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder returns the Microsoft Teams definition builder with the supplied operator config applied
func Builder(cfg Config) definition.Builder {
	return definition.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Slug:        Slug,
				Family:      "microsoft",
				DisplayName: "Microsoft Teams",
				Description: "Send notification messages to Microsoft Teams channels via Microsoft Graph.",
				Category:    "collab",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/microsoft_teams/overview",
				Labels:      map[string]string{"vendor": "microsoft", "product": "teams"},
				Active:      false,
				Visible:     true,
			},
			OperatorConfig: &types.OperatorConfigRegistration{
				Schema: providerkit.SchemaFrom[Config](),
			},
			UserInput: &types.UserInputRegistration{
				Schema: providerkit.SchemaFrom[UserInput](),
			},
			CredentialRegistrations: []types.CredentialRegistration{
				{
					Ref:         teamsCredential,
					Name:        "Microsoft Teams Credential",
					Description: "Auth-managed credential slot used by the Microsoft Teams client in this definition.",
				},
			},
			Auth: &types.AuthRegistration{
				StartPath:    types.DefaultAuthStartPath,
				CallbackPath: types.DefaultAuthCompletePath,
				OAuth: &types.OAuthPublicConfig{
					ClientID:    cfg.ClientID,
					AuthURL:     "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
					TokenURL:    "https://login.microsoftonline.com/common/oauth2/v2.0/token",
					RedirectURI: cfg.RedirectURL,
					Scopes: []string{
						"https://graph.microsoft.com/User.Read",
						"https://graph.microsoft.com/ChannelMessage.Send",
						"offline_access",
					},
				},
				ClientSecret: cfg.ClientSecret,
			},
			Clients: []types.ClientRegistration{
				{
					Ref:            TeamsClient.ID(),
					CredentialRefs: []types.CredentialRef{teamsCredential},
					Description:    "Microsoft Graph API client",
					Build:          Client{}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Call Graph /me to verify Teams access",
					Topic:       HealthDefaultOperation.Topic(Slug),
					ClientRef:   TeamsClient.ID(),
					Handle:      HealthCheck{}.Handle(),
				},
				{
					Name:         MessageSendOperation.Name(),
					Description:  "Send a Teams channel message via Microsoft Graph",
					Topic:        MessageSendOperation.Topic(Slug),
					ClientRef:    TeamsClient.ID(),
					ConfigSchema: providerkit.SchemaFrom[MessageOperationInput](),
					Handle:       MessageSend{}.Handle(),
				},
			},
		}, nil
	})
}
