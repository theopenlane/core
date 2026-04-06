package slacknotify

import (
	"bytes"
	"context"
	"html/template"
	"os"
	"strings"

	"github.com/theopenlane/utils/slack"

	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/slacktemplates"
)

// SlackConfig defines the runtime configuration for Slack notifications
type SlackConfig struct {
	// WebhookURL is the endpoint to send messages to
	WebhookURL string
	// NewSubscriberMessageFile is an optional path to a bespoke Slack template for new subscriber notifications
	NewSubscriberMessageFile string
	// NewUserMessageFile is an optional path to a bespoke Slack template for new user notifications
	NewUserMessageFile string
}

var cfg SlackConfig

// SetConfig replaces the active Slack notification configuration
func SetConfig(c SlackConfig) {
	cfg = c
}

// GetConfig returns the active Slack notification configuration
func GetConfig() SlackConfig {
	return cfg
}

// NotificationsEnabled reports whether a Slack webhook is configured
func NotificationsEnabled() bool {
	return cfg.WebhookURL != ""
}

// SendNotification posts a plain Slack message when webhook notifications are configured
func SendNotification(ctx context.Context, message string) error {
	if !NotificationsEnabled() || message == "" {
		return nil
	}

	if ctx == nil {
		ctx = context.Background()
	}

	return slack.New(cfg.WebhookURL).Post(ctx, &slack.Payload{Text: message})
}

// SendNotificationWithEmail renders and posts a Slack message using explicit email and context values
func SendNotificationWithEmail(ctx context.Context, email, overrideFile, embeddedTemplate string) error {
	if !NotificationsEnabled() {
		return nil
	}

	if ctx == nil {
		ctx = context.Background()
	}

	tmpl, err := LoadTemplate(ctx, overrideFile, embeddedTemplate)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, struct{ Email string }{Email: email}); err != nil {
		logx.FromContext(ctx).Debug().Msg("failed to execute slack template")
		return err
	}

	return slack.New(cfg.WebhookURL).Post(ctx, &slack.Payload{Text: buf.String()})
}

// githubAppInstallSlackTemplateData captures template data for GitHub App install notifications
type githubAppInstallSlackTemplateData struct {
	GitHubOrganization         string
	GitHubAccountType          string
	OpenlaneOrganization       string
	OpenlaneOrganizationID     string
	ShowOpenlaneOrganizationID bool
}

// RenderGitHubAppInstallMessage renders the GitHub App installation Slack message
func RenderGitHubAppInstallMessage(githubOrg, githubAccountType, openlaneOrgName, openlaneOrgID string) (string, error) {
	if githubOrg == "" {
		githubOrg = "unknown"
	}

	if openlaneOrgName == "" {
		openlaneOrgName = openlaneOrgID
	}

	tmpl, err := LoadTemplate(context.Background(), "", slacktemplates.GitHubAppInstallName)
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

// SubscriberTemplateOverride returns the subscriber template override path if configured
func SubscriberTemplateOverride() string {
	return strings.TrimSpace(cfg.NewSubscriberMessageFile)
}

// UserTemplateOverride returns the user template override path if configured
func UserTemplateOverride() string {
	return strings.TrimSpace(cfg.NewUserMessageFile)
}

// LoadTemplate resolves either an override template from disk or a bundled template
func LoadTemplate(ctx context.Context, fileOverride, embeddedTemplate string) (*template.Template, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if fileOverride != "" {
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
