package handlers

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/definitions/catalog"
	"github.com/theopenlane/core/internal/integrations/definitions/githubapp"
	"github.com/theopenlane/core/internal/integrations/registry"
	integrationsruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/internal/slacknotify"
)

// newGitHubAppRuntimeForTest builds a minimal integration runtime for unit tests.
// Pass a non-nil cfg to register the GitHub App definition. Pass nil for tests that
// need to verify behaviour when no definition is registered.
func newGitHubAppRuntimeForTest(t *testing.T, cfg *githubapp.Config) *integrationsruntime.Runtime {
	t.Helper()

	reg := registry.New()
	if cfg != nil {
		require.NoError(t, definition.RegisterAll(reg, githubapp.Builder(*cfg)))
	}

	return integrationsruntime.NewForTesting(reg)
}

// TestValidateGitHubAppConfig verifies required configuration errors and success cases.
func TestValidateGitHubAppConfig(t *testing.T) {
	cases := []struct {
		name    string
		cfg     *githubapp.Config
		wantErr error
	}{
		{
			name:    "provider not configured",
			cfg:     nil,
			wantErr: ErrProviderDisabled,
		},
		{
			name:    "missing app slug",
			cfg:     &githubapp.Config{},
			wantErr: errGitHubAppNotConfigured,
		},
		{
			name: "valid",
			cfg:  &githubapp.Config{AppSlug: "openlane"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := &Handler{IntegrationsRuntime: newGitHubAppRuntimeForTest(t, tc.cfg)}
			if tc.cfg != nil {
				h.IntegrationsConfig = catalog.Config{GitHubApp: *tc.cfg}
			}
			err := h.validateGitHubAppConfig()
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestGitHubAppInstallURL verifies install URL construction.
func TestGitHubAppInstallURL(t *testing.T) {
	cfg := &githubapp.Config{AppSlug: "openlane"}
	h := &Handler{
		IntegrationsRuntime: newGitHubAppRuntimeForTest(t, cfg),
		IntegrationsConfig:  catalog.Config{GitHubApp: *cfg},
	}

	installURL, err := h.githubAppInstallURL("state-value")
	assert.NoError(t, err)

	parsed, err := url.Parse(installURL)
	assert.NoError(t, err)
	assert.Equal(t, "github.com", parsed.Host)
	assert.Equal(t, "/apps/openlane/installations/new", parsed.Path)
	assert.Equal(t, "state-value", parsed.Query().Get("state"))
}

// TestGitHubAppInstallURLMissingSlug verifies that an unregistered definition returns provider disabled.
func TestGitHubAppInstallURLMissingSlug(t *testing.T) {
	// no provider registered → Definition() returns !ok → ErrProviderDisabled
	h := &Handler{IntegrationsRuntime: newGitHubAppRuntimeForTest(t, nil)}

	_, err := h.githubAppInstallURL("state")
	assert.ErrorIs(t, err, ErrProviderDisabled)
}

// TestRenderGitHubAppInstallSlackMessage verifies Slack notification message formatting.
func TestRenderGitHubAppInstallSlackMessage(t *testing.T) {
	msg, err := slacknotify.RenderGitHubAppInstallMessage("acme-corp", "Organization", "Acme", "org_123")
	assert.NoError(t, err)
	assert.Contains(t, msg, "GitHub organization: acme-corp")
	assert.Contains(t, msg, "GitHub account type: Organization")
	assert.Contains(t, msg, "Openlane organization: Acme (org_123)")

	msg, err = slacknotify.RenderGitHubAppInstallMessage("", "", "", "org_123")
	assert.NoError(t, err)
	assert.Contains(t, msg, "GitHub organization: unknown")
	assert.Contains(t, msg, "Openlane organization: org_123")
}
