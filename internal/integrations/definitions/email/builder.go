package email

import (
	"fmt"

	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// Builder returns the email definition builder with the supplied runtime config applied
func Builder(cfg *RuntimeEmailConfig, devMode bool) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		def := types.Definition{
			ID:          DefinitionID.ID(),
			Family:      "email",
			DisplayName: "Email",
			Description: "Send templated transactional and campaign emails via resend.",
			Category:    "messaging",
			DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/email/overview",
			Tags:        []string{"email", "messaging", "notifications"},
			Active:      true,
			Visible:     true,
			CredentialRegistrations: []types.CredentialRegistration{
				emailCredentialRef.Registration(types.CredentialRegistration{
					Name:        "Email Provider Credential",
					Description: "API key and provider selection for email delivery",
				}),
			},
			HealthCheck: emailClientRef.HealthCheck(checkHealth),
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef: emailCredentialRef.ID(),
					Name:          "Email Provider API Key",
					Description:   "Configure email delivery using an API key for resend, sendgrid, or postmark",
				},
			},
			Clients: []types.ClientRegistration{
				emailClientRef.Registration(buildCustomerClient, types.ClientRegistration{
					Description: "Email provider client via newman",
				}),
			},
			UserInput: userInput.Registration(),
			Operations: append(AllEmailOperations(),
				SendEmailOp.Description("Send a single templated email").Registration(DefinitionID),
				SendCampaignOp.Description("Dispatch an email campaign").Registration(DefinitionID),
				SendQuestionnaireCampaignOp.Description("Dispatch a questionnaire campaign").Registration(DefinitionID),
				RecurringCampaignOp.Description("Dispatch due recurring campaigns").Registration(DefinitionID),
				TrustCenterNotificationOp.Description("Notify trust center subscribers about stable posts and subprocessor changes").Registration(DefinitionID),
			),
		}

		deliveryHandler := ResendDeliveryEvent{}.Handle

		def.Webhooks = []types.WebhookRegistration{
			resendWebhookRef.Registration(types.WebhookRegistration{
				StaticRoute:  "/email/webhook",
				SecretSource: func() string { return cfg.ResendSecret },
				Verify:       ResendWebhook{Secret: cfg.ResendSecret}.Verify,
				Event:        ResendWebhook{}.Event,
				Events: []types.WebhookEventRegistration{
					resendEmailSentEvent.Registration(DefinitionID, types.WebhookEventRegistration{Handle: deliveryHandler}),
					resendEmailDeliveredEvent.Registration(DefinitionID, types.WebhookEventRegistration{Handle: deliveryHandler}),
					resendEmailOpenedEvent.Registration(DefinitionID, types.WebhookEventRegistration{Handle: deliveryHandler}),
					resendEmailClickedEvent.Registration(DefinitionID, types.WebhookEventRegistration{Handle: deliveryHandler}),
					resendEmailBouncedEvent.Registration(DefinitionID, types.WebhookEventRegistration{Handle: deliveryHandler}),
					resendEmailFailedEvent.Registration(DefinitionID, types.WebhookEventRegistration{Handle: deliveryHandler}),
				},
			}),
		}

		if len(cfg.Social) == 0 {
			cfg.Social = DefaultSocial
		}

		if devMode || cfg.Provisioned() {
			config, err := jsonx.ToRawMessage(cfg)
			if err != nil {
				return types.Definition{}, fmt.Errorf("%w: %w", ErrClientBuildFailed, err)
			}

			def.RuntimeIntegration = &types.RuntimeIntegrationRegistration{
				Schema: jsonx.SchemaFrom[RuntimeEmailConfig](),
				Config: config,
				Build:  runtimeClientBuilder(devMode && !cfg.Provisioned()),
			}
		}

		return def, nil
	})
}
