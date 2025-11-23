package providers

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/config"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Provider ensures all concrete providers satisfy the shared contract
type Provider interface {
	types.Provider
}

// Builder creates provider instances from specs
type Builder interface {
	// Type returns the provider type this builder handles
	Type() types.ProviderType
	// Build constructs the provider instance from the spec
	Build(ctx context.Context, spec config.ProviderSpec) (Provider, error)
}

// BuilderFunc adapts a function to the Builder interface
type BuilderFunc struct {
	ProviderType types.ProviderType
	BuildFunc    func(ctx context.Context, spec config.ProviderSpec) (Provider, error)
}

// Type returns the provider identifier handled by the builder
func (f BuilderFunc) Type() types.ProviderType {
	return f.ProviderType
}

// Build constructs the provider using the wrapped function
func (f BuilderFunc) Build(ctx context.Context, spec config.ProviderSpec) (Provider, error) {
	if f.BuildFunc == nil {
		return nil, ErrRelyingPartyInit
	}
	return f.BuildFunc(ctx, spec)
}
