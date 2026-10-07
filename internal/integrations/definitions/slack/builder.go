package slack

import (
	"fmt"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/auth"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// Builder returns the Slack definition builder with the operator and runtime config
func Builder(cfg Config, runtime *RuntimeSlackConfig, devMode bool) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		def := types.Definition{
			ID:          DefinitionID.ID(),
			Family:      "Slack",
			DisplayName: "Slack",
			Description: "Integrate with Slack to verify workspace posture and send operational or compliance notifications.",
			Category:    "collaboration",
			DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/slack/overview",
			Tags:        []string{"messaging", "directory"},
			Active:      true,
			Visible:     true,
			OperatorConfig: &types.OperatorConfigRegistration{
				Schema: jsonx.SchemaFrom[Config](),
			},
			UserInput:    userInput.Registration(),
			Installation: installation.Registration(),
			Connections: []types.Connector{
				oauthConnection.
					Name("Slack OAuth").
					Description("Connect your Slack workspace via OAuth").
					Authenticates(auth.OAuthRegistration(auth.OAuthRegistrationOptions[slackCred]{
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
					})).
					Provides(oauthClient).
					Verified(verify[slackCred]).
					Disconnects("Removes the stored OAuth credential from Openlane. To fully revoke access, remove the Openlane app from your Slack workspace under Administration > Manage apps.", disconnectWorkspace(cfg)),
				botTokenConnection.
					Name("Slack Bot Token").
					Description("Connect your Slack workspace using a bot token from a custom Slack app.").
					Provides(botTokenClient).
					Verified(verify[slackBotTokenCred]).
					Disconnects("Removes the stored bot token from Openlane. To fully revoke access, delete or regenerate the token in your Slack app under OAuth & Permissions.", nil),
			},
			Operations: append(AllSlackSystemMessages(),
				MessageSendOp.Description("Send a Slack message via chat.postMessage").Registration(),
				types.OperationRefOf[DirectorySync]().
					Ingests(runDirectorySync).
					Policy(types.ExecutionPolicy{Reconcile: true, Snapshot: true}).
					Ingest(types.IngestContract{Schema: entityops.SchemaDirectoryAccount.Name}).
					Permissions(scopes...).
					Description("Collect workspace users as directory accounts").
					Registration(),
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
