package githuboauth

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	githubDefaultAuthURL  = "https://github.com/login/oauth/authorize"
	githubDefaultTokenURL = "https://github.com/login/oauth/access_token"
)

var githubScopes = []string{
	"read:user",
	"user:email",
	"repo",
	"security_events",
	"admin:repo_hook",
}

// startInstallAuth starts the GitHub OAuth install flow
func (d *def) startInstallAuth(ctx context.Context, _ json.RawMessage) (types.AuthStartResult, error) {
	return providerkit.StartOAuthFlow(ctx, providerkit.OAuthFlowConfig{
		ClientID:     d.cfg.ClientID,
		ClientSecret: d.cfg.ClientSecret,
		AuthURL:      githubAuthURL(d.cfg.BaseURL),
		TokenURL:     githubTokenURL(d.cfg.BaseURL),
		RedirectURL:  d.cfg.RedirectURL,
		Scopes:       githubScopes,
	})
}

// completeInstallAuth completes the GitHub OAuth install flow
func (d *def) completeInstallAuth(ctx context.Context, state json.RawMessage, input json.RawMessage) (types.AuthCompleteResult, error) {
	return providerkit.CompleteOAuthFlow(ctx, providerkit.OAuthFlowConfig{
		ClientID:     d.cfg.ClientID,
		ClientSecret: d.cfg.ClientSecret,
		AuthURL:      githubAuthURL(d.cfg.BaseURL),
		TokenURL:     githubTokenURL(d.cfg.BaseURL),
		RedirectURL:  d.cfg.RedirectURL,
		Scopes:       githubScopes,
	}, state, input)
}

// githubAuthURL returns the authorization URL, substituting a custom base for GitHub Enterprise
func githubAuthURL(baseURL string) string {
	if b := strings.TrimRight(baseURL, "/"); b != "" {
		return b + "/login/oauth/authorize"
	}

	return githubDefaultAuthURL
}

// githubTokenURL returns the token URL, substituting a custom base for GitHub Enterprise
func githubTokenURL(baseURL string) string {
	if b := strings.TrimRight(baseURL, "/"); b != "" {
		return b + "/login/oauth/access_token"
	}

	return githubDefaultTokenURL
}
