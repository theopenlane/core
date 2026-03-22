package microsoftteams

import (
	"github.com/theopenlane/core/internal/integrations/auth"
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
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:       teamsCredential,
					Name:                "Microsoft Teams OAuth",
					Description:         "Authenticate with Microsoft Graph to send Teams channel messages.",
					CredentialRefs:      []types.CredentialRef{teamsCredential},
					ClientRefs:          []types.ClientID{TeamsClient.ID()},
					ValidationOperation: HealthDefaultOperation.Name(),
					Installation:        Installation.Registration(),
					Auth: auth.OAuthRegistration(auth.OAuthRegistrationOptions[teamsCred]{
						CredentialRef: teamsCredential,
						Config: auth.OAuthConfig{
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
						TokenView: func(cred teamsCred) (*types.TokenView, error) {
							return &types.TokenView{
								AccessToken: cred.AccessToken,
								ExpiresAt:   cred.Expiry,
							}, nil
						},
						EncodeCredentialError: ErrCredentialEncode,
						DecodeCredentialError: ErrCredentialDecode,
					}),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: teamsCredential,
						Name:          "Disconnect Microsoft Teams OAuth",
						Description:   "Remove the persisted Microsoft Teams OAuth credential and disconnect this installation from Openlane.",
					},
				},
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
