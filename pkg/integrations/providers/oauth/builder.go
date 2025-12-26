package oauth

import (
	"context"

	"github.com/theopenlane/common/integrations/config"
	"github.com/theopenlane/common/integrations/types"
	"github.com/theopenlane/core/pkg/integrations/providers"
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
