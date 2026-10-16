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
			ID:           DefinitionID.ID(),
			Family:       "email",
			DisplayName:  "Email",
			Description:  "Send templated transactional and campaign emails via Resend or Postmark.",
			Category:     "messaging",
			DocsURL:      "https://docs.theopenlane.io/docs/platform/integrations/email/overview",
			Tags:         []string{"email", "messaging", "notifications"},
			Active:       true,
			Visible:      true,
			Installation: installation.Registration(),
			Connections: []types.Connector{
				resendConnection.
					Name("Resend API Key").
					Description("Configure email delivery through Resend using an API key").
					Recommended().
					Provides(customerClientBuilder(resendSender)).
					Verified(verifySender[ResendCredential](providerResend)),
				postmarkConnection.
					Name("Postmark Server Token").
					Description("Configure email delivery through Postmark using a server API token").
					Provides(customerClientBuilder(postmarkSender)).
					Verified(verifySender[PostmarkCredential](providerPostmark)),
			},
			UserInput: userInput.Registration(),
			Operations: append(AllEmailOperations(),
				SendEmailOp.Description("Send a single templated email").Registration(),
				SendCampaignOp.Description("Dispatch an email campaign").Registration(),
				SendQuestionnaireCampaignOp.Description("Dispatch a questionnaire campaign").Registration(),
				RecurringCampaignOp.Description("Dispatch due recurring campaigns").Registration(),
				TrustCenterNotificationOp.Description("Notify trust center subscribers about stable posts and subprocessor changes").Registration(),
			),
		}

		deliveryHandler := ResendDeliveryEvent{}.Handle
		postmarkHandler := PostmarkDeliveryEvent{}.Handle

		resendWebhook := resendWebhookRef.Registration(types.WebhookRegistration{
			StaticRoute: "/email/webhook",
			Verify:      ResendWebhook{Secret: cfg.Resend.WebhookSecret}.Verify,
			Event:       ResendWebhook{}.Event,
			Events: []types.WebhookEventRegistration{
				resendEmailSentEvent.Registration(types.WebhookEventRegistration{Handle: deliveryHandler}),
				resendEmailDeliveredEvent.Registration(types.WebhookEventRegistration{Handle: deliveryHandler}),
				resendEmailOpenedEvent.Registration(types.WebhookEventRegistration{Handle: deliveryHandler}),
				resendEmailClickedEvent.Registration(types.WebhookEventRegistration{Handle: deliveryHandler}),
				resendEmailBouncedEvent.Registration(types.WebhookEventRegistration{Handle: deliveryHandler}),
				resendEmailFailedEvent.Registration(types.WebhookEventRegistration{Handle: deliveryHandler}),
			},
		})

		postmarkWebhook := postmarkWebhookRef.Registration(types.WebhookRegistration{
			StaticRoute: "/email/webhook/postmark",
			Verify:      PostmarkWebhook{Secret: cfg.Postmark.WebhookSecret}.Verify,
			Event:       PostmarkWebhook{}.Event,
			Events: []types.WebhookEventRegistration{
				postmarkDeliveryEvent.Registration(types.WebhookEventRegistration{Handle: postmarkHandler}),
				postmarkBounceEvent.Registration(types.WebhookEventRegistration{Handle: postmarkHandler}),
				postmarkSpamComplaintEvent.Registration(types.WebhookEventRegistration{Handle: postmarkHandler}),
				postmarkOpenEvent.Registration(types.WebhookEventRegistration{Handle: postmarkHandler}),
				postmarkClickEvent.Registration(types.WebhookEventRegistration{Handle: postmarkHandler}),
			},
		})

		// only the selected provider's inbound route is registered
		switch cfg.Provider {
		case providerResend:
			def.Webhooks = []types.WebhookRegistration{resendWebhook}
		case providerPostmark:
			def.Webhooks = []types.WebhookRegistration{postmarkWebhook}
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
