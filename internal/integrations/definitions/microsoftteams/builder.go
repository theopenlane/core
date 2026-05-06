package microsoftteams

import (
	"github.com/theopenlane/core/internal/integrations/auth"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder returns the Microsoft Teams definition builder with the supplied operator config applied
func Builder(cfg Config) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Family:      "Microsoft",
				DisplayName: "Microsoft Teams",
				Description: "Send notification messages to Microsoft Teams channels via Microsoft Graph.",
				Category:    "collaboration",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/microsoft_teams/overview",
				Tags:        []string{"messaging"},
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
					Ref:         teamsCredential.ID(),
					Name:        "Microsoft Teams Credential",
					Description: "OAuth credential used to send messages to Microsoft Teams channels.",
					Schema:      teamsCredentialSchema,
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:       teamsCredential.ID(),
					Name:                "Microsoft Teams OAuth",
					Description:         "Connect your Microsoft Teams workspace using OAuth.",
					CredentialRefs:      []types.CredentialSlotID{teamsCredential.ID()},
					ClientRefs:          []types.ClientID{teamsClient.ID()},
					ValidationOperation: healthCheckOperation.Name(),
					Integration:         installation.Registration(),
					Auth: auth.OAuthRegistration(auth.OAuthRegistrationOptions[teamsCred]{
						CredentialRef: teamsCredential,
						Config: auth.OAuthConfig{ //nolint:gosec
							ClientID:     cfg.ClientID,
							ClientSecret: cfg.ClientSecret,
							AuthURL:      "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
							TokenURL:     "https://login.microsoftonline.com/common/oauth2/v2.0/token",
							RedirectURL:  cfg.RedirectURL,
							Scopes: []string{
								"https://graph.microsoft.com/User.Read",
								"https://graph.microsoft.com/ChannelMessage.Send",
								"offline_access",
							},
						},
						Material: func(material auth.OAuthMaterial) (teamsCred, error) {
							return teamsCred{
								AccessToken:  material.AccessToken,
								RefreshToken: material.RefreshToken,
								Expiry:       material.Expiry,
							}, nil
						},
						EncodeCredentialError: ErrCredentialEncode,
					}),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: teamsCredential.ID(),
						Description:   "Removes the stored OAuth credential from Openlane. To fully revoke access, remove the Openlane app from your Azure Entra ID enterprise applications.",
					},
				},
			},
			Clients: []types.ClientRegistration{
				{
					Ref:            teamsClient.ID(),
					CredentialRefs: []types.CredentialSlotID{teamsCredential.ID()},
					Description:    "Microsoft Graph API client",
					Build:          Client{}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:         healthCheckOperation.Name(),
					Description:  "Call Graph /me to verify Teams access",
					Topic:        DefinitionID.OperationTopic(healthCheckOperation.Name()),
					ClientRef:    teamsClient.ID(),
					Policy:       types.ExecutionPolicy{Inline: true},
					ConfigSchema: healthCheckSchema,
					Handle:       HealthCheck{}.Handle(),
				},
				{
					Name:         MessageSendOp.Name(),
					Description:  "Send a Teams channel message via Microsoft Graph",
					Topic:        DefinitionID.OperationTopic(MessageSendOp.Name()),
					ClientRef:    teamsClient.ID(),
					ConfigSchema: messageSendSchema,
					Handle:       MessageSend{}.Handle(),
				},
			},
		}, nil
	})
}
