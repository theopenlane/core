package hooks

import (
	"bytes"
	"context"
	"html/template"
	"os"
	"strings"

	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/slacktemplates"
	"github.com/theopenlane/utils/slack"
)

// SlackConfig defines the runtime configuration for Slack notifications emitted by listeners
type SlackConfig struct {
	// WebhookURL is the endpoint to send messages to
	WebhookURL string
	// NewSubscriberMessageFile is an optional path to a bespoke Slack template for new subscriber notifications
	NewSubscriberMessageFile string
	// NewUserMessageFile is an optional path to a bespoke Slack template for new user notifications
	NewUserMessageFile string
}

var slackCfg SlackConfig

// SetSlackConfig replaces the active Slack notification configuration
func SetSlackConfig(cfg SlackConfig) {
	slackCfg = cfg
}

// sendSlackNotificationWithEmail renders and posts a Slack message using explicit email and context values.
func sendSlackNotificationWithEmail(ctx context.Context, email, overrideFile, embeddedTemplate string) error {
	if slackCfg.WebhookURL == "" {
		return nil
	}

	if ctx == nil {
		ctx = context.Background()
	}

	tmpl, err := loadSlackTemplate(ctx, overrideFile, embeddedTemplate)
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	if err := tmpl.Execute(&buf, struct{ Email string }{Email: email}); err != nil {
		logx.FromContext(ctx).Debug().Msg("failed to execute slack template")
		return err
	}

	return slack.New(slackCfg.WebhookURL).Post(ctx, &slack.Payload{Text: buf.String()})
}

// loadSlackTemplate resolves either an override template from disk or a bundled template for Slack notifications
func loadSlackTemplate(ctx context.Context, fileOverride, embeddedTemplate string) (*template.Template, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if fileOverride != "" {
		// Allow operators to provide a bespoke template from disk; fall back to the embedded
		// version when it is absent so environments without customisation still behave.
		b, err := os.ReadFile(fileOverride)
		if err != nil {
			logx.FromContext(ctx).Debug().Msg("failed to read slack template")

			return nil, err
		}

		t, err := template.New("slack").Parse(string(b))
		if err != nil {
			logx.FromContext(ctx).Debug().Msg("failed to parse slack template")

			return nil, err
		}

		return t, nil
	}

	t, err := template.ParseFS(slacktemplates.Templates, embeddedTemplate)
	if err != nil {
		logx.FromContext(ctx).Debug().Msg("failed to parse embedded slack template")

		return nil, err
	}

	return t, nil
}

// subscriberTemplateOverride returns the subscriber template override if configured.
func subscriberTemplateOverride() string {
	return strings.TrimSpace(slackCfg.NewSubscriberMessageFile)
}

// userTemplateOverride returns the user template override if configured.
func userTemplateOverride() string {
	return strings.TrimSpace(slackCfg.NewUserMessageFile)
}
