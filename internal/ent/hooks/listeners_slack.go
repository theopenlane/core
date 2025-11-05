package hooks

import (
	"bytes"
	"html/template"
	"os"

	"entgo.io/ent"

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
}

var slackCfg SlackConfig

// SetSlackConfig replaces the active Slack notification configuration
func SetSlackConfig(cfg SlackConfig) {
	slackCfg = cfg
}

// handleSubscriberMutation processes subscriber mutations and publishes Slack notifications when a record is created
func handleSubscriberMutation(ctx *soiree.EventContext, payload *MutationPayload) error {
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
func handleUserMutation(ctx *soiree.EventContext, payload *MutationPayload) error {
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
	if slackCfg.WebhookURL == "" {
		return nil
	}

	email, _ := ctx.PropertyString("email")

	tmpl, err := loadSlackTemplate(ctx, overrideFile, embeddedTemplate)
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	if err := tmpl.Execute(&buf, struct{ Email string }{Email: email}); err != nil {
		logx.FromContext(ctx.Context()).Debug().Msg("failed to execute slack template")
		return err
	}

	return slack.New(slackCfg.WebhookURL).Post(ctx.Context(), &slack.Payload{Text: buf.String()})
}

// loadSlackTemplate resolves either an override template from disk or a bundled template for Slack notifications
func loadSlackTemplate(ctx *soiree.EventContext, fileOverride, embeddedTemplate string) (*template.Template, error) {
	if fileOverride != "" {
		// Allow operators to provide a bespoke template from disk; fall back to the embedded
		// version when it is absent so environments without customisation still behave.
		b, err := os.ReadFile(fileOverride)
		if err != nil {
			logx.FromContext(ctx.Context()).Debug().Msg("failed to read slack template")

			return nil, err
		}

		t, err := template.New("slack").Parse(string(b))
		if err != nil {
			logx.FromContext(ctx.Context()).Debug().Msg("failed to parse slack template")

			return nil, err
		}

		return t, nil
	}

	t, err := template.ParseFS(slacktemplates.Templates, embeddedTemplate)
	if err != nil {
		logx.FromContext(ctx.Context()).Debug().Msg("failed to parse embedded slack template")

		return nil, err
	}

	return t, nil
}
