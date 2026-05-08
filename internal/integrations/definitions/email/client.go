package email

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/newman"
	"github.com/theopenlane/newman/providers/mock"
	"github.com/theopenlane/newman/providers/postmark"
	"github.com/theopenlane/newman/providers/resend"
	"github.com/theopenlane/newman/providers/sendgrid"
	"github.com/theopenlane/newman/scrubber"
)

// Client wraps a newman.EmailSender with the RuntimeEmailConfig
// needed for template rendering and branding
type Client struct {
	// Sender is the configured email provider client
	Sender newman.EmailSender
	// Config holds branding and URL configuration for template rendering
	Config RuntimeEmailConfig
}

// buildRuntimeClient constructs an Client from marshaled RuntimeEmailConfig
func buildRuntimeClient(_ context.Context, config json.RawMessage) (any, error) {
	var cfg RuntimeEmailConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrClientBuildFailed, err)
	}

	sender, err := buildSender(cfg.Provider, cfg.APIKey)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrClientBuildFailed, err)
	}

	return &Client{
		Sender: sender,
		Config: cfg,
	}, nil
}

// buildCustomerClient constructs an Client from resolved credentials and user input
func buildCustomerClient(_ context.Context, req types.ClientBuildRequest) (any, error) {
	cred, _, err := emailCredentialRef.Resolve(req.Credentials)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrClientBuildFailed, err)
	}

	sender, senderErr := buildSender(cred.Provider, cred.APIKey)
	if senderErr != nil {
		return nil, fmt.Errorf("%w: %w", ErrClientBuildFailed, senderErr)
	}

	var userInput UserInput
	if req.Integration != nil {
		if err := jsonx.UnmarshalIfPresent(req.Integration.Config.ClientConfig, &userInput); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrClientBuildFailed, err)
		}
	}

	return &Client{
		Sender: sender,
		Config: userInput.ToRuntimeConfig(),
	}, nil
}

// EmailHTMLScrubber is the shared scrubber instance used by email providers for render-time
// HTML sanitization, preserving email-safe layout elements while stripping dangerous content
var EmailHTMLScrubber = scrubber.NewPolicyScrubber(scrubber.WithEmailDefaults())

// EmailScrubber returns the shared HTML scrubber instance used for email
// content sanitization. This is the canonical sanitization policy for both
// storage-time and render-time email HTML processing
func EmailScrubber() scrubber.Scrubber { //nolint:revive
	return EmailHTMLScrubber
}

// ProviderMock is the provider name for the mock email sender used in tests
const ProviderMock = "mock"

// buildSender constructs a newman.EmailSender for the given provider
func buildSender(provider string, apiKey string) (newman.EmailSender, error) {
	switch provider {
	case "resend":
		return resend.New(apiKey, resend.WithHTMLScrubber(EmailHTMLScrubber))
	case "sendgrid":
		return sendgrid.New(apiKey, sendgrid.WithHTMLScrubber(EmailHTMLScrubber))
	case "postmark":
		return postmark.New(apiKey, postmark.WithHTMLScrubber(EmailHTMLScrubber))
	case ProviderMock:
		return mock.New("")
	default:
		return nil, fmt.Errorf("%w: %s", ErrProviderNotSupported, provider)
	}
}
