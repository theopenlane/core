package oauth

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/spec"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder returns a providers.Builder that constructs an OAuth provider for the given provider type
func Builder(providerType types.ProviderType, opts ...Option) providers.Builder {
	return providerkit.Builder(providerType, func(_ context.Context, s spec.ProviderSpec) (types.Provider, error) {
		return New(s, opts...)
	})
}
