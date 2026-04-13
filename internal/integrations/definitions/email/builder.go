package email

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// Builder returns the email definition builder with the supplied runtime config applied.
// When cfg.Provisioned() is true, a RuntimeIntegration is included for system-send.
// Customer registrations (credentials, connections, clients, user input) are always present
func Builder(cfg *RuntimeEmailConfig) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		def := types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID.ID(),
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
				Schema: emailUserInputSchema,
			},
			MutationListeners: []types.MutationListenerRegistration{
				{
					Name:          "email.campaign.active",
					SchemaType:    "Campaign",
					OperationName: sendCampaignOp.Name(),
					Handle:        handleCampaignActivation,
				},
			},
			Operations: append(AllEmailOperations(), types.OperationRegistration{
				Name:         sendCampaignOp.Name(),
				Description:  "Dispatch a full email campaign",
				Topic:        definitionID.OperationTopic(sendCampaignOp.Name()),
				ClientRef:    emailClientRef.ID(),
				ConfigSchema: sendCampaignSchema,
				Policy:       types.ExecutionPolicy{Reconcile: true},
				Handle:       handleSendCampaign,
			}),
		}

		if cfg.Provisioned() {
			runtimeEmailRef.SetConfig(cfg)

			marshaledConfig, err := runtimeEmailRef.MarshalConfig()
			if err != nil {
				return types.Definition{}, fmt.Errorf("%w: %w", ErrClientBuildFailed, err)
			}

			def.RuntimeIntegration = &types.RuntimeIntegrationRegistration{
				Ref:    runtimeEmailRef.ID(),
				Schema: runtimeEmailSchema,
				Config: marshaledConfig,
				Build:  buildRuntimeClient,
			}
		}

		return def, nil
	})
}

// buildRuntimeClient constructs an EmailClient from marshaled RuntimeEmailConfig
func buildRuntimeClient(_ context.Context, config json.RawMessage) (any, error) {
	var cfg RuntimeEmailConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrClientBuildFailed, err)
	}

	sender, err := buildSender(cfg.Provider, cfg.APIKey)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrClientBuildFailed, err)
	}

	return &EmailClient{
		Sender: sender,
		Config: cfg,
	}, nil
}

// buildCustomerClient constructs an EmailClient from resolved credentials and user input
func buildCustomerClient(_ context.Context, req types.ClientBuildRequest) (any, error) {
	cred, _, err := emailCredentialRef.Resolve(req.Credentials)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrClientBuildFailed, err)
	}

	sender, senderErr := buildSender(cred.Provider, cred.APIKey)
	if senderErr != nil {
		return nil, fmt.Errorf("%w: %w", ErrClientBuildFailed, senderErr)
	}

	var userInput EmailUserInput
	if req.Integration != nil {
		if err := jsonx.UnmarshalIfPresent(req.Integration.Config.ClientConfig, &userInput); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrClientBuildFailed, err)
		}
	}

	return &EmailClient{
		Sender: sender,
		Config: userInput.ToRuntimeConfig(),
	}, nil
}
