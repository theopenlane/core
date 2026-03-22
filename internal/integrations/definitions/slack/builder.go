package slack

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/auth"
	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder returns the Slack definition builder with the supplied operator config applied
func Builder(cfg Config) definition.Builder {
	return definition.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Slug:        Slug,
				Family:      "slack",
				DisplayName: "Slack",
				Description: "Integrate with Slack to verify workspace posture and send operational or compliance notifications.",
				Category:    "collab",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/slack/overview",
				Labels:      map[string]string{"vendor": "slack"},
				Active:      true,
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
					Ref:         slackCredential,
					Name:        "Slack Credential",
					Description: "Auth-managed credential slot used by the Slack client in this definition.",
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:       slackCredential,
					Name:                "Slack OAuth",
					Description:         "Authenticate with Slack using an installed bot token.",
					CredentialRefs:      []types.CredentialRef{slackCredential},
					ClientRefs:          []types.ClientID{SlackClient.ID()},
					ValidationOperation: HealthDefaultOperation.Name(),
					Installation:        Installation.Registration(),
					Auth: auth.OAuthRegistration(auth.OAuthRegistrationOptions[slackCred]{
						CredentialRef: slackCredential,
						Config: auth.OAuthConfig{
							ClientID:     cfg.ClientID,
							ClientSecret: cfg.ClientSecret,
							AuthURL:      "https://slack.com/oauth/v2/authorize",
							TokenURL:     "https://slack.com/api/oauth.v2.access",
							RedirectURL:  cfg.RedirectURL,
							Scopes: []string{
								"chat:write",
								"chat:write.public",
								"chat:write.customize",
								"channels:read",
								"groups:read",
								"team:read",
								"users:read",
								"users:read.email",
								"users:profile:read",
							},
						},
						Material: func(material auth.OAuthMaterial) (slackCred, error) {
							return slackCred{
								AccessToken:  material.AccessToken,
								RefreshToken: material.RefreshToken,
								Expiry:       material.Expiry,
							}, nil
						},
						TokenView: func(cred slackCred) (*types.TokenView, error) {
							return &types.TokenView{
								AccessToken: cred.AccessToken,
								ExpiresAt:   cred.Expiry,
							}, nil
						},
						EncodeCredentialError: ErrCredentialEncode,
						DecodeCredentialError: ErrCredentialDecode,
					}),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: slackCredential,
						Name:          "Disconnect Slack OAuth",
						Description:   "Remove the persisted Slack OAuth credential and disconnect this workspace installation from Openlane.",
					},
				},
			},
			Clients: []types.ClientRegistration{
				{
					Ref:            SlackClient.ID(),
					CredentialRefs: []types.CredentialRef{slackCredential},
					Description:    "Slack Web API client",
					Build:          Client{}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Call auth.test to ensure the Slack token is valid and scoped correctly",
					Topic:       HealthDefaultOperation.Topic(Slug),
					ClientRef:   SlackClient.ID(),
					Handle:      HealthCheck{}.Handle(),
				},
				{
					Name:        TeamInspectOperation.Name(),
					Description: "Collect workspace metadata via team.info for posture analysis",
					Topic:       TeamInspectOperation.Topic(Slug),
					ClientRef:   SlackClient.ID(),
					Handle:      TeamInspect{}.Handle(),
				},
				{
					Name:         ChannelsListOperation.Name(),
					Description:  "List Slack channels available for use as notification destinations",
					Topic:        ChannelsListOperation.Topic(Slug),
					ClientRef:    SlackClient.ID(),
					ConfigSchema: providerkit.SchemaFrom[ChannelsListOperationInput](),
					Policy:       types.ExecutionPolicy{Inline: true},
					Handle:       ChannelsList{}.Handle(),
				},
				{
					Name:         MessageSendOperation.Name(),
					Description:  "Send a Slack message via chat.postMessage",
					Topic:        MessageSendOperation.Topic(Slug),
					ClientRef:    SlackClient.ID(),
					ConfigSchema: providerkit.SchemaFrom[MessageOperationInput](),
					Handle:       MessageSend{}.Handle(),
				},
				{
					Name:        DirectorySyncOperation.Name(),
					Description: "Collect workspace users as directory accounts",
					Topic:       DirectorySyncOperation.Topic(Slug),
					ClientRef:   SlackClient.ID(),
					Ingest: []types.IngestContract{
						{
							Schema: integrationgenerated.IntegrationMappingSchemaDirectoryAccount,
						},
					},
					IngestHandle: DirectorySync{}.IngestHandle(),
				},
			},
			Mappings: slackMappings(),
		}, nil
	})
}
