package email

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/newman"
	"github.com/theopenlane/newman/providers/mock"
	"github.com/theopenlane/newman/providers/postmark"
	"github.com/theopenlane/newman/providers/resend"
	"github.com/theopenlane/newman/scrubber"
)

// Client wraps a newman.EmailSender with the RuntimeEmailConfig needed for template rendering and branding
type Client struct {
	// Sender is the configured email provider client
	Sender newman.EmailSender
	// Config holds branding and URL configuration for template rendering
	Config RuntimeEmailConfig
}

// runtimeClientBuilder returns a build function that constructs a Client from marshaled RuntimeEmailConfig, writing MIME files to TestDir instead of calling the provider API when devMode is true
func runtimeClientBuilder(devMode bool) func(context.Context, json.RawMessage) (any, error) {
	return func(_ context.Context, config json.RawMessage) (any, error) {
		var cfg RuntimeEmailConfig
		if err := json.Unmarshal(config, &cfg); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrClientBuildFailed, err)
		}

		sender, err := runtimeSender(cfg, devMode)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrClientBuildFailed, err)
		}

		return &Client{
			Sender: sender,
			Config: cfg,
		}, nil
	}
}

// customerClientBuilder returns a connection client builder that pairs the sender newSender constructs from the decoded credential with the installation user input
func customerClientBuilder[T any](newSender func(T) (newman.EmailSender, error)) func(context.Context, types.ConnectionRequest[T]) (*Client, error) {
	return func(_ context.Context, req types.ConnectionRequest[T]) (*Client, error) {
		sender, err := newSender(req.Credential)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrClientBuildFailed, err)
		}

		var input UserInput
		if req.Integration != nil {
			input, err = userInput.Decode(req.Integration)
			if err != nil {
				return nil, fmt.Errorf("%w: %w", ErrClientBuildFailed, err)
			}
		}

		return &Client{
			Sender: sender,
			Config: input.ToRuntimeConfig(),
		}, nil
	}
}

// resendSender constructs the Resend sender for a customer-provisioned credential
func resendSender(cred ResendCredential) (newman.EmailSender, error) {
	return resend.New(cred.APIKey, resend.WithHTMLScrubber(EmailHTMLScrubber))
}

// postmarkSender constructs the Postmark sender for a customer-provisioned credential, applying its stream and tracking options
func postmarkSender(cred PostmarkCredential) (newman.EmailSender, error) {
	opts := []postmark.Option{postmark.WithHTMLScrubber(EmailHTMLScrubber)}

	if cred.MessageStream != "" {
		opts = append(opts, postmark.WithMessageStream(cred.MessageStream))
	}

	if cred.TrackOpens {
		opts = append(opts, postmark.WithTrackOpens())
	}

	if cred.TrackLinks != "" {
		opts = append(opts, postmark.WithTrackLinks(cred.TrackLinks))
	}

	return postmark.New(cred.ServerToken, opts...)
}

// EmailHTMLScrubber is the shared scrubber instance used by email providers for render-time HTML sanitization, preserving email-safe layout elements while stripping dangerous content
var EmailHTMLScrubber = scrubber.NewPolicyScrubber(scrubber.WithEmailDefaults())

// Scrubber returns the shared HTML scrubber instance used for email content sanitization, the canonical policy for both storage-time and render-time processing
func Scrubber() scrubber.Scrubber {
	return EmailHTMLScrubber
}

const (
	// ProviderMock is the provider name for the mock email sender used in tests
	ProviderMock = "mock"
	// providerResend is the provider name recorded for Resend delivery
	providerResend = "resend"
	// providerPostmark is the provider name recorded for Postmark delivery
	providerPostmark = "postmark"
)

// runtimeSender constructs the sender for the provider cfg selects, from that provider's own credential block; devMode routes every provider to the mock writer that outputs MIME files to TestDir instead of sending
func runtimeSender(cfg RuntimeEmailConfig, devMode bool) (newman.EmailSender, error) {
	if devMode {
		return mock.New(cfg.TestDir)
	}

	switch cfg.Provider {
	case providerResend:
		return resendSender(cfg.Resend.ResendCredential)
	case providerPostmark:
		return postmarkSender(cfg.Postmark.PostmarkCredential)
	case ProviderMock:
		return mock.New("")
	default:
		return nil, fmt.Errorf("%w: %s", ErrProviderNotSupported, cfg.Provider)
	}
}
