package googleworkspace

import (
	"context"

	"github.com/theopenlane/core/common/integrations/config"
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/oauth"
)

// TypeGoogleWorkspace identifies the Google Workspace provider
const TypeGoogleWorkspace = types.ProviderType("googleworkspace")

// oauthProvider wraps oauth.Provider and implements DirectoryAccountMappingProvider.
type oauthProvider struct {
	*oauth.Provider
}

// DefaultDirectoryAccountMappings returns the built-in directory account mapping specs for Google Workspace.
func (p *oauthProvider) DefaultDirectoryAccountMappings() map[string]types.MappingSpec {
	return googleWorkspaceDirectoryAccountMappings()
}

// Builder returns the Google Workspace provider builder.
func Builder() providers.Builder {
	return providers.BuilderFunc{
		ProviderType: TypeGoogleWorkspace,
		BuildFunc: func(_ context.Context, spec config.ProviderSpec) (providers.Provider, error) {
			base, err := oauth.New(spec,
				oauth.WithOperations(googleWorkspaceOperations()),
				oauth.WithClientDescriptors(googleWorkspaceClientDescriptors()),
			)
			if err != nil {
				return nil, err
			}

			return &oauthProvider{Provider: base}, nil
		},
	}
}
