package github

import (
	"context"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/oauth"
	"github.com/theopenlane/core/internal/integrations/spec"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	// TypeGitHub identifies the GitHub OAuth provider
	TypeGitHub = types.ProviderType("github")
)

// oauthProvider wraps oauth.Provider and implements types.MappingProvider.
type oauthProvider struct {
	*oauth.Provider
}

// DefaultMappings returns the built-in ingest mapping registrations for GitHub OAuth providers.
func (p *oauthProvider) DefaultMappings() []types.MappingRegistration {
	return githubDefaultMappings()
}

// Builder returns the GitHub OAuth provider builder with the supplied operator config applied.
func Builder(cfg Config) providers.Builder {
	return providers.BuilderFunc{
		ProviderType: TypeGitHub,
		SpecFunc:     githubSpec,
		BuildFunc: func(_ context.Context, s spec.ProviderSpec) (types.Provider, error) {
			if s.OAuth != nil && cfg.ClientID != "" {
				s.OAuth.ClientID = cfg.ClientID
				s.OAuth.ClientSecret = cfg.ClientSecret
			}

			base, err := oauth.New(s,
				oauth.WithOperations(githubOperations()),
				oauth.WithClientDescriptors(githubClientDescriptors(TypeGitHub)),
			)
			if err != nil {
				return nil, err
			}

			return &oauthProvider{Provider: base}, nil
		},
	}
}

// githubSpec returns the static provider specification for the GitHub OAuth provider.
func githubSpec() spec.ProviderSpec {
	return spec.ProviderSpec{
		Name:             "github",
		DisplayName:      "GitHub",
		Category:         "code",
		AuthType:         types.AuthKindOAuth2,
		AuthStartPath:    "/v1/integrations/oauth/start",
		AuthCallbackPath: "/v1/integrations/oauth/callback",
		Active:           lo.ToPtr(true),
		Visible:          lo.ToPtr(false),
		LogoURL:          "",
		DocsURL:          "https://docs.theopenlane.io/docs/platform/integrations/github/overview",
		OAuth: &spec.OAuthSpec{
			AuthURL:     "https://github.com/login/oauth/authorize",
			TokenURL:    "https://github.com/login/oauth/access_token",
			Scopes:      []string{"read:user", "user:email", "repo", "security_events", "admin:repo_hook"},
			RedirectURI: "http://localhost:17608/v1/integrations/oauth/callback",
		},
		UserInfo: &spec.UserInfoSpec{
			URL:               "https://api.github.com/user",
			Method:            "GET",
			AuthStyle:         "token",
			IDPath:            "id",
			EmailPath:         "email",
			LoginPath:         "login",
			SecondaryEmailURL: "https://api.github.com/user/emails",
		},
		Persistence: &spec.PersistenceSpec{
			StoreRefreshToken: true,
		},
		Description: "Collect GitHub repository metadata and security alerts (Dependabot, code scanning, and secret scanning) for exposure management.",
	}
}
