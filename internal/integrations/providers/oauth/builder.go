package oauth

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/config"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder returns a providers.Builder that constructs OAuth providers for the given provider type
func Builder(provider types.ProviderType) providers.Builder {
	return providers.BuilderFunc{
		ProviderType: provider,
		BuildFunc: func(ctx context.Context, spec config.ProviderSpec) (providers.Provider, error) {
			return New(ctx, spec)
		},
	}
}
