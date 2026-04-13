package email

import (
	"fmt"

	"github.com/theopenlane/newman"
	"github.com/theopenlane/newman/providers/postmark"
	"github.com/theopenlane/newman/providers/resend"
	"github.com/theopenlane/newman/providers/sendgrid"
)

// EmailClient wraps a newman.EmailSender with the RuntimeEmailConfig
// needed for template rendering and branding
type EmailClient struct {
	// Sender is the configured email provider client
	Sender newman.EmailSender
	// Config holds branding and URL configuration for template rendering
	Config RuntimeEmailConfig
}

// buildSender constructs a newman.EmailSender for the given provider
func buildSender(provider string, apiKey string) (newman.EmailSender, error) {
	switch provider {
	case "resend":
		return resend.New(apiKey)
	case "sendgrid":
		return sendgrid.New(apiKey)
	case "postmark":
		return postmark.New(apiKey)
	default:
		return nil, fmt.Errorf("%w: %s", ErrProviderNotSupported, provider)
	}
}
