package slack

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/auth"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// Builder returns the Slack definition builder with the supplied operator and runtime config applied
func Builder(cfg Config, runtime *RuntimeSlackConfig, devMode bool) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		def := types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Family:      "Slack",
				DisplayName: "Slack",
				Description: "Integrate with Slack to verify workspace posture and send operational or compliance notifications.",
				Category:    "collaboration",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/slack/overview",
				Tags:        []string{"messaging", "directory"},
				Active:      true,
				Visible:     true,
			},
			OperatorConfig: &types.OperatorConfigRegistration{
				Schema: jsonx.SchemaFrom[Config](),
			},
			UserInput:    userInput.Registration(),
			HealthCheck:  slackClient.HealthCheck(checkHealth),
			Installation: installation.Registration(),
			CredentialRegistrations: []types.CredentialRegistration{
				slackCredential.Registration(types.CredentialRegistration{
					Name:        "Slack OAuth Credential",
					Description: "OAuth credential used to access the Slack workspace",
				}),
				slackBotTokenCredential.Registration(types.CredentialRegistration{
					Name:        "Slack Bot Token",
					Description: "User-provisioned bot token from a custom Slack app",
				}),
			},
			Connections: []types.ConnectionRegistration{
				slackOAuthConnection.Registration(types.ConnectionRegistration{
					Name:        "Slack OAuth",
					Description: "Connect your Slack workspace via OAuth",
					Auth: auth.OAuthRegistration(auth.OAuthRegistrationOptions[slackCred]{
						CredentialRef: slackCredential,
						Config: auth.OAuthConfig{ //nolint:gosec
							ClientID:     cfg.ClientID,
							ClientSecret: cfg.ClientSecret,
							AuthURL:      "https://slack.com/oauth/v2/authorize",
							TokenURL:     "https://slack.com/api/oauth.v2.access",
							RedirectURL:  cfg.RedirectURL,
							Scopes:       scopes,
						},
						Material: func(material auth.OAuthMaterial) (slackCred, error) {
							return slackCred{
								AccessToken:  material.AccessToken,
								RefreshToken: material.RefreshToken,
								Expiry:       material.Expiry,
							}, nil
						},
						EncodeCredentialError: ErrCredentialEncode,
					}),
					Disconnect: &types.DisconnectRegistration{
						Description: "Removes the stored OAuth credential from Openlane. To fully revoke access, remove the Openlane app from your Slack workspace under Administration > Manage apps.",
						Disconnect: func(_ context.Context, req types.DisconnectRequest) (types.DisconnectResult, error) {
							teamID, err := disconnectTeamID(req)
							if err != nil {
								return types.DisconnectResult{}, err
							}

							details, err := jsonx.ToRawMessage(disconnectDetails{
								TeamID: teamID,
							})
							if err != nil {
								return types.DisconnectResult{}, ErrInstallationMetadataEncode
							}

							url := getManageURL(teamID, cfg.AppID)

							return types.DisconnectResult{
								RedirectURL: url,
								Message:     "Uninstall the Openlane Slack App in Slack to finish disconnecting this integration.",
								Details:     details,
							}, nil
						},
					},
				}),
				slackBotTokenConnection.Registration(types.ConnectionRegistration{
					Name:        "Slack Bot Token",
					Description: "Connect your Slack workspace using a bot token from a custom Slack app.",
					Disconnect: &types.DisconnectRegistration{
						Description: "Removes the stored bot token from Openlane. To fully revoke access, delete or regenerate the token in your Slack app under OAuth & Permissions.",
					},
				}),
			},
			Clients: []types.ClientRegistration{
				slackClient.Registration(Client{}.Build, types.ClientRegistration{
					Description: "Unified Slack client wrapping the Web API and system-notification transports",
				}),
			},
			Operations: append(AllSlackSystemMessages(),
				MessageSendOp.Registration(DefinitionID, types.OperationRegistration{
					Description:         "Send a Slack message via chat.postMessage",
					RequiredPermissions: scopes,
				}),
				directorySyncOperation.Registration(DefinitionID, types.OperationRegistration{
					Description: "Collect workspace users as directory accounts",
					Policy:      types.ExecutionPolicy{Reconcile: true, Snapshot: true},
					Ingest: []types.IngestContract{
						{
							Schema: entityops.SchemaDirectoryAccount.Name,
						},
					},
					RequiredPermissions: scopes,
				}),
			),
			Mappings: []types.MappingRegistration{
				{
					Schema: entityops.SchemaDirectoryAccount.Name,
					Spec: types.MappingOverride{
						FilterExpr: "true",
						MapExpr:    mapExprDirectoryAccount,
					},
				},
			},
		}

		if runtime != nil && (devMode || runtime.Provisioned()) {
			config, err := jsonx.ToRawMessage(runtime)
			if err != nil {
				return types.Definition{}, fmt.Errorf("%w: %w", ErrClientBuildFailed, err)
			}

			def.RuntimeIntegration = &types.RuntimeIntegrationRegistration{
				Schema: jsonx.SchemaFrom[RuntimeSlackConfig](),
				Config: config,
				Build:  runtimeSlackClientBuilder(devMode && !runtime.Provisioned()),
			}
		}

		return def, nil
	})
}

var scopes = []string{
	"chat:write",
	"chat:write.public",
	"chat:write.customize",
	"channels:read",
	"groups:read",
	"team:read",
	"users:read",
	"users:read.email",
	"users.profile:read",
}
