package slack

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/auth"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder returns the Slack definition builder with the supplied operator config applied
func Builder(cfg Config) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID.ID(),
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
					Ref:         slackCredential.ID(),
					Name:        "Slack Credential",
					Description: "Auth-managed credential slot used by the Slack client in this definition.",
					Schema:      slackCredentialSchema,
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:       slackCredential.ID(),
					Name:                "Slack OAuth",
					Description:         "Authenticate with Slack using an installed bot token.",
					CredentialRefs:      []types.CredentialSlotID{slackCredential.ID()},
					ClientRefs:          []types.ClientID{slackClient.ID()},
					ValidationOperation: healthCheckOperation.Name(),
					Installation:        installation.Registration(),
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
						CredentialRef: slackCredential.ID(),
						Name:          "Disconnect Slack OAuth",
						Description:   "Remove the persisted Slack OAuth credential and disconnect this workspace installation from Openlane.",
					},
				},
			},
			Clients: []types.ClientRegistration{
				{
					Ref:            slackClient.ID(),
					CredentialRefs: []types.CredentialSlotID{slackCredential.ID()},
					Description:    "Slack Web API client",
					Build:          Client{}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:         healthCheckOperation.Name(),
					Description:  "Call auth.test to ensure the Slack token is valid and scoped correctly",
					Topic:        types.OperationTopic(definitionID.ID(), healthCheckOperation.Name()),
					ClientRef:    slackClient.ID(),
					Policy:       types.ExecutionPolicy{Inline: true},
					ConfigSchema: healthCheckSchema,
					Handle:       HealthCheck{}.Handle(),
				},
				{
					Name:         messageSendOperation.Name(),
					Description:  "Send a Slack message via chat.postMessage",
					Topic:        types.OperationTopic(definitionID.ID(), messageSendOperation.Name()),
					ClientRef:    slackClient.ID(),
					ConfigSchema: messageSendSchema,
					Handle:       MessageSend{}.Handle(),
				},
				{
					Name:         directorySyncOperation.Name(),
					Description:  "Collect workspace users as directory accounts",
					Topic:        types.OperationTopic(definitionID.ID(), directorySyncOperation.Name()),
					ClientRef:    slackClient.ID(),
					ConfigSchema: directorySyncSchema,
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
