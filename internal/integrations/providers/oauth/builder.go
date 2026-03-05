package oauth

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/config"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder returns a providers.Builder that constructs OAuth providers for the given provider type
func Builder(provider types.ProviderType, opts ...ProviderOption) providers.Builder {
	return providerkit.Builder(provider, func(_ context.Context, spec config.ProviderSpec) (providers.Provider, error) {
		return New(spec, opts...)
	})
}
