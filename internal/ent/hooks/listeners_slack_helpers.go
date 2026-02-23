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

// githubAppInstallSlackTemplateData captures template data for GitHub App install notifications.
type githubAppInstallSlackTemplateData struct {
	// GitHubOrganization is the installed GitHub account login.
	GitHubOrganization string
	// GitHubAccountType is the GitHub account type (User or Organization).
	GitHubAccountType string
	// OpenlaneOrganization is the Openlane organization display name.
	OpenlaneOrganization string
	// OpenlaneOrganizationID is the Openlane organization identifier.
	OpenlaneOrganizationID string
	// ShowOpenlaneOrganizationID controls whether to append the organization ID in Slack text.
	ShowOpenlaneOrganizationID bool
}

// SetSlackConfig replaces the active Slack notification configuration
func SetSlackConfig(cfg SlackConfig) {
	slackCfg = cfg
}

// SlackNotificationsEnabled reports whether a Slack webhook is configured.
func SlackNotificationsEnabled() bool {
	return slackCfg.WebhookURL != ""
}

// SendSlackNotification posts a plain Slack message when webhook notifications are configured.
func SendSlackNotification(ctx context.Context, message string) error {
	if !SlackNotificationsEnabled() {
		return nil
	}

	if message == "" {
		return nil
	}

	if ctx == nil {
		ctx = context.Background()
	}

	return slack.New(slackCfg.WebhookURL).Post(ctx, &slack.Payload{Text: message})
}

// RenderGitHubAppInstallSlackMessage renders the GitHub App installation Slack message.
func RenderGitHubAppInstallSlackMessage(githubOrg, githubAccountType, openlaneOrgName, openlaneOrgID string) (string, error) {
	if githubOrg == "" {
		githubOrg = "unknown"
	}

	if openlaneOrgName == "" {
		openlaneOrgName = openlaneOrgID
	}

	tmpl, err := loadSlackTemplate(context.Background(), "", slacktemplates.GitHubAppInstallName)
	if err != nil {
		return "", err
	}

	data := githubAppInstallSlackTemplateData{
		GitHubOrganization:         githubOrg,
		GitHubAccountType:          githubAccountType,
		OpenlaneOrganization:       openlaneOrgName,
		OpenlaneOrganizationID:     openlaneOrgID,
		ShowOpenlaneOrganizationID: openlaneOrgID != "" && openlaneOrgID != openlaneOrgName,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// sendSlackNotificationWithEmail renders and posts a Slack message using explicit email and context values.
func sendSlackNotificationWithEmail(ctx context.Context, email, overrideFile, embeddedTemplate string) error {
	if !SlackNotificationsEnabled() {
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
