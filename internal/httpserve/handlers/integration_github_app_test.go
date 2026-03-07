package handlers

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/theopenlane/core/internal/ent/hooks"
	integrationruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/utils/rout"
)

// TestValidateGitHubAppConfig verifies required configuration errors and success cases.
func TestValidateGitHubAppConfig(t *testing.T) {
	valid := integrationruntime.GitHubAppConfig{
		Enabled:       true,
		AppSlug:       "openlane",
		AppID:         "12345",
		PrivateKey:    "private-key",
		WebhookSecret: "secret",
	}

	cases := []struct {
		name      string
		cfg       integrationruntime.GitHubAppConfig
		wantErr   error
		wantField string
	}{
		{
			name:    "disabled",
			cfg:     integrationruntime.GitHubAppConfig{},
			wantErr: ErrProviderDisabled,
		},
		{
			name:      "missing slug",
			cfg:       integrationruntime.GitHubAppConfig{Enabled: true, AppID: valid.AppID, PrivateKey: valid.PrivateKey, WebhookSecret: valid.WebhookSecret},
			wantField: "appSlug",
		},
		{
			name:      "missing app id",
			cfg:       integrationruntime.GitHubAppConfig{Enabled: true, AppSlug: valid.AppSlug, PrivateKey: valid.PrivateKey, WebhookSecret: valid.WebhookSecret},
			wantField: "appId",
		},
		{
			name:      "missing private key",
			cfg:       integrationruntime.GitHubAppConfig{Enabled: true, AppSlug: valid.AppSlug, AppID: valid.AppID, WebhookSecret: valid.WebhookSecret},
			wantField: "privateKey",
		},
		{
			name:      "missing webhook secret",
			cfg:       integrationruntime.GitHubAppConfig{Enabled: true, AppSlug: valid.AppSlug, AppID: valid.AppID, PrivateKey: valid.PrivateKey},
			wantField: "webhookSecret",
		},
		{
			name: "valid",
			cfg:  valid,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			h := &Handler{IntegrationRuntime: integrationruntime.NewConfigOnly(tc.cfg, integrationruntime.OAuthConfig{})}
			err := h.validateGitHubAppConfig()
			switch {
			case tc.wantField != "":
				assert.Error(t, err)
				var fieldErr *rout.FieldError
				assert.ErrorAs(t, err, &fieldErr)
				assert.Equal(t, tc.wantField, fieldErr.Field)
				assert.ErrorIs(t, err, rout.ErrMissingField)
			case tc.wantErr != nil:
				assert.ErrorIs(t, err, tc.wantErr)
			default:
				assert.NoError(t, err)
			}
		})
	}
}

// TestGitHubAppInstallURL verifies install URL construction and missing slug errors.
func TestGitHubAppInstallURL(t *testing.T) {
	h := &Handler{IntegrationRuntime: integrationruntime.NewConfigOnly(
		integrationruntime.GitHubAppConfig{AppSlug: "openlane"},
		integrationruntime.OAuthConfig{},
	)}

	installURL, err := h.githubAppInstallURL("state-value")
	assert.NoError(t, err)

	parsed, err := url.Parse(installURL)
	assert.NoError(t, err)
	assert.Equal(t, "github.com", parsed.Host)
	assert.Equal(t, "/apps/openlane/installations/new", parsed.Path)
	assert.Equal(t, "state-value", parsed.Query().Get("state"))
}

// TestGitHubAppInstallURLMissingSlug verifies missing slug errors use field helpers.
func TestGitHubAppInstallURLMissingSlug(t *testing.T) {
	h := &Handler{IntegrationRuntime: integrationruntime.NewConfigOnly(
		integrationruntime.GitHubAppConfig{},
		integrationruntime.OAuthConfig{},
	)}

	_, err := h.githubAppInstallURL("state")
	assert.Error(t, err)
	var fieldErr *rout.FieldError
	assert.ErrorAs(t, err, &fieldErr)
	assert.Equal(t, "appSlug", fieldErr.Field)
	assert.ErrorIs(t, err, rout.ErrMissingField)
}

// TestRenderGitHubAppInstallSlackMessage verifies Slack notification message formatting.
func TestRenderGitHubAppInstallSlackMessage(t *testing.T) {
	msg, err := hooks.RenderGitHubAppInstallSlackMessage("acme-corp", "Organization", "Acme", "org_123")
	assert.NoError(t, err)
	assert.Contains(t, msg, "GitHub organization: acme-corp")
	assert.Contains(t, msg, "GitHub account type: Organization")
	assert.Contains(t, msg, "Openlane organization: Acme (org_123)")

	msg, err = hooks.RenderGitHubAppInstallSlackMessage("", "", "", "org_123")
	assert.NoError(t, err)
	assert.Contains(t, msg, "GitHub organization: unknown")
	assert.Contains(t, msg, "Openlane organization: org_123")
}
