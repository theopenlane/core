package handlers

import (
	"context"
	"net/url"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/core/internal/ent/hooks"
	integrationconfig "github.com/theopenlane/core/internal/integrations/config"
	githubprovider "github.com/theopenlane/core/internal/integrations/providers/github"
	"github.com/theopenlane/core/internal/integrations/registry"
	integrationruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/utils/rout"
)

// newGitHubAppRuntimeForTest builds a minimal integration runtime backed by a registry
// containing the given GitHub App provider spec. If spec.Name is empty, no provider is registered.
func newGitHubAppRuntimeForTest(t *testing.T, spec integrationconfig.ProviderSpec) *integrationruntime.Runtime {
	t.Helper()

	ctx := context.Background()
	reg, err := registry.NewRegistry(ctx, nil)
	require.NoError(t, err)

	if spec.Name != "" {
		require.NoError(t, reg.UpsertProvider(ctx, spec, githubprovider.AppBuilder()))
	}

	return integrationruntime.NewFromRegistry(reg)
}

// TestValidateGitHubAppConfig verifies required configuration errors and success cases.
func TestValidateGitHubAppConfig(t *testing.T) {
	cases := []struct {
		name    string
		spec    integrationconfig.ProviderSpec
		wantErr error
	}{
		{
			name: "provider disabled",
			spec: integrationconfig.ProviderSpec{
				Name:     string(githubprovider.TypeGitHubApp),
				Active:   lo.ToPtr(false),
				AuthType: types.AuthKindGitHubApp,
			},
			wantErr: ErrProviderDisabled,
		},
		{
			name: "missing credentials",
			spec: integrationconfig.ProviderSpec{
				Name:     string(githubprovider.TypeGitHubApp),
				Active:   lo.ToPtr(true),
				AuthType: types.AuthKindGitHubApp,
				GitHubApp: &integrationconfig.GitHubAppSpec{
					AppSlug: "openlane",
					// AppID, PrivateKey, WebhookSecret intentionally absent
				},
			},
			wantErr: errGitHubAppNotConfigured,
		},
		{
			name: "valid",
			spec: integrationconfig.ProviderSpec{
				Name:     string(githubprovider.TypeGitHubApp),
				Active:   lo.ToPtr(true),
				AuthType: types.AuthKindGitHubApp,
				GitHubApp: &integrationconfig.GitHubAppSpec{
					AppSlug:       "openlane",
					AppID:         "12345",
					PrivateKey:    "private-key",
					WebhookSecret: "secret",
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := &Handler{IntegrationRuntime: newGitHubAppRuntimeForTest(t, tc.spec)}
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
	spec := integrationconfig.ProviderSpec{
		Name:     string(githubprovider.TypeGitHubApp),
		Active:   lo.ToPtr(true),
		AuthType: types.AuthKindGitHubApp,
		GitHubApp: &integrationconfig.GitHubAppSpec{
			AppSlug: "openlane",
		},
	}
	h := &Handler{IntegrationRuntime: newGitHubAppRuntimeForTest(t, spec)}

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
	// no provider registered → gitHubAppSpec returns ok=false → MissingField("appSlug")
	h := &Handler{IntegrationRuntime: newGitHubAppRuntimeForTest(t, integrationconfig.ProviderSpec{})}

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
