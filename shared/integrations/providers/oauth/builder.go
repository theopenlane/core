package oauth

import (
	"context"

	"github.com/theopenlane/shared/integrations/config"
	"github.com/theopenlane/shared/integrations/providers"
	"github.com/theopenlane/shared/integrations/types"
)

// Builder returns a providers.Builder that constructs OAuth providers for the given provider type
func Builder(provider types.ProviderType, opts ...ProviderOption) providers.Builder {
	return providers.BuilderFunc{
		ProviderType: provider,
		BuildFunc: func(ctx context.Context, spec config.ProviderSpec) (providers.Provider, error) {
			if spec.OAuth == nil {
				return nil, nil
			}

			return New(ctx, spec, opts...)
		},
	}
}
