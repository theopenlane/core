package providerkit

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/config"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/types"
)

// BuildFunc constructs a provider from a provider spec
type BuildFunc func(ctx context.Context, spec config.ProviderSpec) (providers.Provider, error)

// Builder returns a providers.Builder for the given provider type and build function
func Builder(provider types.ProviderType, build BuildFunc) providers.Builder {
	return providers.BuilderFunc{
		ProviderType: provider,
		BuildFunc:    build,
	}
}

// ValidateAuthType checks that an optional auth type declaration matches the expected auth kind
func ValidateAuthType(spec config.ProviderSpec, expected types.AuthKind, mismatch error) error {
	if spec.AuthType != "" && spec.AuthType != expected {
		return mismatch
	}

	return nil
}
