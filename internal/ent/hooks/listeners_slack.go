package hooks

import (
	"bytes"
	"context"
	"html/template"
	"os"
	"strings"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/events"
	"github.com/theopenlane/core/pkg/events/soiree"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/slacktemplates"
	"github.com/theopenlane/utils/slack"
)

// SlackConfig defines the runtime configuration for Slack notifications emitted by listeners
type SlackConfig struct {
	// WebhookURL is the endpoint to send messages to
	WebhookURL string
	// NewSubscriberMessageFile is an optional path to a bespoke Slack template for new subscriber notifications (cat memes)
	NewSubscriberMessageFile string
	// NewUserMessageFile is an optional path to a bespoke Slack template for new user notifications (welcome messages)
	NewUserMessageFile string
	// GalaNewSubscriberMessageFile is an optional path to a Gala-specific subscriber template
	GalaNewSubscriberMessageFile string
	// GalaNewUserMessageFile is an optional path to a Gala-specific user template
	GalaNewUserMessageFile string
}

var slackCfg SlackConfig

// SetSlackConfig replaces the active Slack notification configuration
func SetSlackConfig(cfg SlackConfig) {
	slackCfg = cfg
}

// handleSubscriberMutation processes subscriber mutations and publishes Slack notifications when a record is created
func handleSubscriberMutation(ctx *soiree.EventContext, payload *events.MutationPayload) error {
	if payload.Operation != ent.OpCreate.String() {
		return nil
	}

	return sendSlackNotification(
		ctx,
		slackCfg.NewSubscriberMessageFile,
		slacktemplates.SubscriberTemplateName,
	)
}

// handleUserMutation processes user mutations and publishes Slack notifications when a record is created
func handleUserMutation(ctx *soiree.EventContext, payload *events.MutationPayload) error {
	if payload.Operation != ent.OpCreate.String() {
		return nil
	}

	return sendSlackNotification(
		ctx,
		slackCfg.NewUserMessageFile,
		slacktemplates.UserTemplateName,
	)
}

// sendSlackNotification renders the desired Slack template and posts the resulting message to the configured webhook
func sendSlackNotification(ctx *soiree.EventContext, overrideFile, embeddedTemplate string) error {
	email := ""
	logCtx := context.Background()
	if ctx != nil {
		email, _ = ctx.PropertyString("email")
		logCtx = ctx.Context()
	}

	return sendSlackNotificationWithEmail(logCtx, email, overrideFile, embeddedTemplate)
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

// galaSubscriberTemplateOverride returns the Gala-specific subscriber template override if configured.
func galaSubscriberTemplateOverride() string {
	override := strings.TrimSpace(slackCfg.GalaNewSubscriberMessageFile)
	if override != "" {
		return override
	}

	return strings.TrimSpace(slackCfg.NewSubscriberMessageFile)
}

// galaUserTemplateOverride returns the Gala-specific user template override if configured.
func galaUserTemplateOverride() string {
	override := strings.TrimSpace(slackCfg.GalaNewUserMessageFile)
	if override != "" {
		return override
	}

	return strings.TrimSpace(slackCfg.NewUserMessageFile)
}
