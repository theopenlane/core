package email

import (
	"fmt"

	"github.com/theopenlane/newman"
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

// emailHTMLScrubber is the shared scrubber instance used by email providers for render-time
// HTML sanitization, preserving email-safe layout elements while stripping dangerous content
var emailHTMLScrubber = scrubber.NewPolicyScrubber(scrubber.WithEmailDefaults())

// buildSender constructs a newman.EmailSender for the given provider
func buildSender(provider string, apiKey string) (newman.EmailSender, error) {
	switch provider {
	case "resend":
		return resend.New(apiKey, resend.WithHTMLScrubber(emailHTMLScrubber))
	case "sendgrid":
		return sendgrid.New(apiKey, sendgrid.WithHTMLScrubber(emailHTMLScrubber))
	case "postmark":
		return postmark.New(apiKey, postmark.WithHTMLScrubber(emailHTMLScrubber))
	default:
		return nil, fmt.Errorf("%w: %s", ErrProviderNotSupported, provider)
	}
}
