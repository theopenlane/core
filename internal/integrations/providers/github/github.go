package github

import (
	"context"

	"github.com/theopenlane/core/common/integrations/config"
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/oauth"
)

// TypeGitHub identifies the GitHub provider
const TypeGitHub = types.ProviderType("github")

// oauthProvider wraps oauth.Provider and implements VulnerabilityMappingProvider.
type oauthProvider struct {
	*oauth.Provider
}

// DefaultVulnerabilityMappings returns the built-in vulnerability mapping specs for the GitHub provider.
func (p *oauthProvider) DefaultVulnerabilityMappings() map[string]types.MappingSpec {
	return githubVulnerabilityMappings()
}

// Builder returns the GitHub provider builder.
func Builder() providers.Builder {
	return providers.BuilderFunc{
		ProviderType: TypeGitHub,
		BuildFunc: func(_ context.Context, spec config.ProviderSpec) (providers.Provider, error) {
			base, err := oauth.New(spec,
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
