package microsoftteams

import (
	"github.com/theopenlane/core/v2/internal/integrations/auth"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
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
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/microsoft_teams/",
				Tags:        []string{"messaging"},
				Active:      false,
				Visible:     true,
			},
			OperatorConfig: &types.OperatorConfigRegistration{
				Schema: jsonx.SchemaFrom[Config](),
			},
			UserInput: &types.UserInputRegistration{
				Schema: jsonx.SchemaFrom[UserInput](),
			},
			CredentialRegistrations: []types.CredentialRegistration{
				teamsCredential.Registration(types.CredentialRegistration{
					Name:        "Microsoft Teams Credential",
					Description: "OAuth credential used to send messages to Microsoft Teams channels.",
				}),
			},
			Connections: []types.ConnectionRegistration{
				teamsConnection.Registration(types.ConnectionRegistration{
					Name:           "Microsoft Teams OAuth",
					Description:    "Connect your Microsoft Teams workspace using OAuth.",
					CredentialRefs: []types.CredentialSlotID{teamsCredential.ID()},
					HealthCheck: &types.HealthCheckRegistration{
						ClientRef: teamsClient.ID(),
						Handle:    HealthCheck{}.Handle(),
					},
					Integration: installation.Registration(),
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
				}),
			},
			Clients: []types.ClientRegistration{
				teamsClient.Registration(Client{}.Build, types.ClientRegistration{
					Description: "Microsoft Graph API client",
				}),
			},
			Operations: []types.OperationRegistration{
				MessageSendOp.Registration(DefinitionID, types.OperationRegistration{
					Description: "Send a Teams channel message via Microsoft Graph",
					Handle:      MessageSend{}.Handle(),
				}),
			},
		}, nil
	})
}
