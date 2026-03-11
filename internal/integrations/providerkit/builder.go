package providerkit

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/spec"
	"github.com/theopenlane/core/internal/integrations/types"
)

// BuildFunc constructs a provider from a provider spec
type BuildFunc func(ctx context.Context, spec spec.ProviderSpec) (types.Provider, error)

// Builder returns a providers.Builder for the given provider type and build function
func Builder(provider types.ProviderType, build BuildFunc) providers.Builder {
	return providers.BuilderFunc{
		ProviderType: provider,
		BuildFunc:    build,
	}
}

// ValidateAuthType checks that an optional auth type declaration matches the expected auth kind.
// Returns mismatch when the spec auth type is set but does not equal expected.
func ValidateAuthType(s spec.ProviderSpec, expected types.AuthKind, mismatch error) error {
	if expected.Normalize() == types.AuthKindUnknown {
		return nil
	}

	kind := s.AuthType.Normalize()
	if kind != expected.Normalize() {
		return mismatch
	}

	return nil
}
