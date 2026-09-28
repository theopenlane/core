package email

import (
	"fmt"

	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
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
				Description: "Send templated transactional and campaign emails via resend.",
				Category:    "messaging",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/email/overview",
				Tags:        []string{"email", "messaging", "notifications"},
				Active:      true,
				Visible:     true,
			},
			CredentialRegistrations: []types.CredentialRegistration{
				emailCredentialRef.Registration(types.CredentialRegistration{
					Name:        "Email Provider Credential",
					Description: "API key and provider selection for email delivery",
					Schema:      emailCredentialRef.Schema(),
				}),
			},
			Connections: []types.ConnectionRegistration{
				emailConnection.Registration(types.ConnectionRegistration{
					Name:           "Email Provider API Key",
					Description:    "Configure email delivery using an API key for resend, sendgrid, or postmark",
					CredentialRefs: []types.CredentialSlotID{emailCredentialRef.ID()},
					HealthCheck: &types.HealthCheckRegistration{
						ClientRef: emailClientRef.ID(),
						Handle:    HealthCheck{}.Handle(),
					},
				}),
			},
			Clients: []types.ClientRegistration{
				emailClientRef.Registration(buildCustomerClient, types.ClientRegistration{
					Description: "Email provider client via newman",
				}),
			},
			UserInput: &types.UserInputRegistration{
				Schema: jsonx.SchemaFrom[UserInput](),
			},
			Operations: append(AllEmailOperations(),
				SendEmailOp.Registration(DefinitionID, types.OperationRegistration{
					Description: "Send a single templated email",
					Policy:      types.ExecutionPolicy{SkipRunRecord: true},
					Handle:      SendEmail{}.Handle(),
				}),
				SendCampaignOp.Registration(DefinitionID, types.OperationRegistration{
					Description: "Dispatch an email campaign",
					Policy:      types.ExecutionPolicy{SkipRunRecord: true},
					Handle:      SendBrandedCampaign{}.Handle(),
				}),
				SendQuestionnaireCampaignOp.Registration(DefinitionID, types.OperationRegistration{
					Description: "Dispatch a questionnaire campaign",
					Policy:      types.ExecutionPolicy{SkipRunRecord: true},
					Handle:      SendQuestionnaireCampaign{}.Handle(),
				}),
				// global singleton sweeps: no client ref, they route dispatches to each
				// org's email installation or the runtime provider
				RecurringCampaignOp.Registration(DefinitionID, types.OperationRegistration{
					Description:         "Dispatch due recurring campaigns",
					Policy:              types.ExecutionPolicy{Scheduled: true, SkipRunRecord: true},
					Handle:              RecurringCampaignSweep{}.Handle(),
					CustomerSelectable:  lo.ToPtr(false),
					SkipDefaultLookback: true,
				}),
				TrustCenterNotificationOp.Registration(DefinitionID, types.OperationRegistration{
					Description:         "Notify trust center subscribers about stable posts and subprocessor changes",
					Policy:              types.ExecutionPolicy{Scheduled: true, SkipRunRecord: true},
					Handle:              TrustCenterNotificationSweep{}.Handle(),
					CustomerSelectable:  lo.ToPtr(false),
					SkipDefaultLookback: true,
				}),
			),
		}

		if cfg.ResendSecret != "" {
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

			def.RuntimeIntegration = lo.ToPtr(runtimeEmailRef.Registration(types.RuntimeIntegrationRegistration{
				Config: marshaledConfig,
				Build:  runtimeClientBuilder(devMode && !cfg.Provisioned()),
			}))
		}

		return def, nil
	})
}
