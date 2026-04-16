package email

import (
	"fmt"

	"github.com/theopenlane/newman"
	"github.com/theopenlane/newman/providers/mock"
	"github.com/theopenlane/newman/providers/postmark"
	"github.com/theopenlane/newman/providers/resend"
	"github.com/theopenlane/newman/providers/sendgrid"
	"github.com/theopenlane/newman/scrubber"
)

// EmailClient wraps a newman.EmailSender with the RuntimeEmailConfig
// needed for template rendering and branding
type EmailClient struct {
	// Sender is the configured email provider client
	Sender newman.EmailSender
	// Config holds branding and URL configuration for template rendering
	Config RuntimeEmailConfig
}

// EmailHTMLScrubber is the shared scrubber instance used by email providers for render-time
// HTML sanitization, preserving email-safe layout elements while stripping dangerous content
var EmailHTMLScrubber = scrubber.NewPolicyScrubber(scrubber.WithEmailDefaults())

// EmailScrubber returns the shared HTML scrubber instance used for email
// content sanitization. This is the canonical sanitization policy for both
// storage-time and render-time email HTML processing
func EmailScrubber() scrubber.Scrubber {
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
