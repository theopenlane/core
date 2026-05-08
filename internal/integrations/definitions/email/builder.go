package email

import (
	"fmt"

	"github.com/resend/resend-go/v3"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder returns the email definition builder with the supplied runtime config applied.
// When devMode is true or cfg.Provisioned() returns true, a RuntimeIntegration is
// included for system-send. In dev mode the sender writes MIME files to cfg.TestDir
// instead of calling the provider API.
// Customer registrations (credentials, connections, clients, user input) are always present
func Builder(cfg *RuntimeEmailConfig, devMode bool) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		def := types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Family:      "email",
				DisplayName: "Email",
				Description: "Send templated transactional and campaign emails via resend, sendgrid, or postmark.",
				Category:    "messaging",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/email/overview",
				Tags:        []string{"email", "messaging", "notifications"},
				Active:      true,
				Visible:     true,
			},
			CredentialRegistrations: []types.CredentialRegistration{
				{
					Ref:         emailCredentialRef.ID(),
					Name:        "Email Provider Credential",
					Description: "API key and provider selection for email delivery",
					Schema:      emailCredentialSchema,
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:  emailCredentialRef.ID(),
					Name:           "Email Provider API Key",
					Description:    "Configure email delivery using an API key for resend, sendgrid, or postmark",
					CredentialRefs: []types.CredentialSlotID{emailCredentialRef.ID()},
					ClientRefs:     []types.ClientID{emailClientRef.ID()},
				},
			},
			Clients: []types.ClientRegistration{
				{
					Ref:            emailClientRef.ID(),
					CredentialRefs: []types.CredentialSlotID{emailCredentialRef.ID()},
					Description:    "Email provider client via newman",
					Build:          buildCustomerClient,
				},
			},
			UserInput: &types.UserInputRegistration{
				Schema: providerkit.SchemaFrom[UserInput](),
			},
			Operations: append(AllEmailOperations(),
				types.OperationRegistration{
					Name:         healthCheckOp.Name(),
					Description:  "Validate the email provider credentials are functional",
					Topic:        DefinitionID.OperationTopic(healthCheckOp.Name()),
					ClientRef:    emailClientRef.ID(),
					ConfigSchema: healthCheckSchema,
					Policy:       types.ExecutionPolicy{Inline: true},
					Handle:       HealthCheck{}.Handle(),
				},
				types.OperationRegistration{
					Name:         SendEmailOp.Name(),
					Description:  "Send a single templated email",
					Topic:        DefinitionID.OperationTopic(SendEmailOp.Name()),
					ClientRef:    emailClientRef.ID(),
					ConfigSchema: sendEmailSchema,
					Policy:       types.ExecutionPolicy{SkipRunRecord: true},
					Handle:       SendEmail{}.Handle(),
				},
				types.OperationRegistration{
					Name:         SendCampaignOp.Name(),
					Description:  "Dispatch an email campaign",
					Topic:        DefinitionID.OperationTopic(SendCampaignOp.Name()),
					ClientRef:    emailClientRef.ID(),
					ConfigSchema: sendBrandedCampaignSchema,
					Policy:       types.ExecutionPolicy{SkipRunRecord: true},
					Handle:       SendBrandedCampaign{}.Handle(),
				},
				types.OperationRegistration{
					Name:         SendQuestionnaireCampaignOp.Name(),
					Description:  "Dispatch a questionnaire campaign",
					Topic:        DefinitionID.OperationTopic(SendQuestionnaireCampaignOp.Name()),
					ClientRef:    emailClientRef.ID(),
					ConfigSchema: sendQuestionnaireCampaignSchema,
					Policy:       types.ExecutionPolicy{SkipRunRecord: true},
					Handle:       SendQuestionnaireCampaign{}.Handle(),
				},
			),
		}

		if cfg.ResendSecret != "" {
			deliveryHandler := ResendDeliveryEvent{}.Handle

			def.Webhooks = []types.WebhookRegistration{
				{
					Name:         resendWebhookRef.Name(),
					StaticRoute:  "/email/webhook",
					SecretSource: func() string { return cfg.ResendSecret },
					Verify:       ResendWebhook{Secret: cfg.ResendSecret}.Verify,
					Event:        ResendWebhook{}.Event,
					Events: []types.WebhookEventRegistration{
						{Name: resend.EventEmailSent, Topic: DefinitionID.WebhookEventTopic(resend.EventEmailSent), Handle: deliveryHandler},
						{Name: resend.EventEmailDelivered, Topic: DefinitionID.WebhookEventTopic(resend.EventEmailDelivered), Handle: deliveryHandler},
						{Name: resend.EventEmailOpened, Topic: DefinitionID.WebhookEventTopic(resend.EventEmailOpened), Handle: deliveryHandler},
						{Name: resend.EventEmailClicked, Topic: DefinitionID.WebhookEventTopic(resend.EventEmailClicked), Handle: deliveryHandler},
						{Name: resend.EventEmailBounced, Topic: DefinitionID.WebhookEventTopic(resend.EventEmailBounced), Handle: deliveryHandler},
						{Name: resend.EventEmailFailed, Topic: DefinitionID.WebhookEventTopic(resend.EventEmailFailed), Handle: deliveryHandler},
					},
				},
			}
		}

		if len(cfg.Social) == 0 {
			cfg.Social = DefaultSocial
		}

		if devMode || cfg.Provisioned() {
			runtimeEmailRef.SetConfig(cfg)

			marshaledConfig, err := runtimeEmailRef.MarshalConfig()
			if err != nil {
				return types.Definition{}, fmt.Errorf("%w: %w", ErrClientBuildFailed, err)
			}

			def.RuntimeIntegration = &types.RuntimeIntegrationRegistration{
				Ref:    runtimeEmailRef.ID(),
				Schema: runtimeEmailSchema,
				Config: marshaledConfig,
				Build:  runtimeClientBuilder(devMode && !cfg.Provisioned()),
			}
		}

		return def, nil
	})
}
