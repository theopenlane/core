package providers

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/spec"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder creates provider instances from specs
type Builder interface {
	// Type returns the provider type this builder constructs
	Type() types.ProviderType
	// Spec returns the static provider specification for this builder
	Spec() spec.ProviderSpec
	// Build constructs the provider instance from the spec
	Build(ctx context.Context, spec spec.ProviderSpec) (types.Provider, error)
}

// BuilderFunc adapts a function pair to the Builder interface
type BuilderFunc struct {
	// ProviderType identifies the provider type this builder constructs
	ProviderType types.ProviderType
	// SpecFunc returns the static provider specification for this builder
	SpecFunc func() spec.ProviderSpec
	// BuildFunc is the function that constructs the provider instance from the spec
	BuildFunc func(ctx context.Context, spec spec.ProviderSpec) (types.Provider, error)
}

// Type returns the provider identifier handled by the builder
func (f BuilderFunc) Type() types.ProviderType {
	return f.ProviderType
}

// Spec returns the static provider specification for this builder
func (f BuilderFunc) Spec() spec.ProviderSpec {
	if f.SpecFunc == nil {
		return spec.ProviderSpec{}
	}

	return f.SpecFunc()
}

// Build constructs the provider using the wrapped function
func (f BuilderFunc) Build(ctx context.Context, s spec.ProviderSpec) (types.Provider, error) {
	if f.BuildFunc == nil {
		return nil, ErrBuilderNil
	}

	provider, err := f.BuildFunc(ctx, s)
	if err != nil {
		return nil, err
	}

	if provider == nil {
		return nil, ErrProviderNil
	}

	return provider, nil
}

// BaseProvider stores shared provider metadata and provides default implementations
// for the non-auth parts of types.Provider, types.OperationProvider, and types.ClientProvider
type BaseProvider struct {
	// ProviderType is the unique identifier for this provider (e.g. "github", "slack")
	ProviderType types.ProviderType
	// Caps is a set of capability flags for this provider
	Caps types.ProviderCapabilities
	// Ops is a list of operations published by this provider
	Ops []types.OperationDescriptor
	// Clients is a list of client descriptors published by this provider
	Clients []types.ClientDescriptor
}

// NewBaseProvider constructs a BaseProvider with shared metadata
func NewBaseProvider(providerType types.ProviderType, caps types.ProviderCapabilities, ops []types.OperationDescriptor, clients []types.ClientDescriptor) BaseProvider {
	return BaseProvider{
		ProviderType: providerType,
		Caps:         caps,
		Ops:          ops,
		Clients:      clients,
	}
}

// Type returns the provider identifier
func (p *BaseProvider) Type() types.ProviderType {
	return p.ProviderType
}

// Capabilities returns the capability flags for this provider
func (p *BaseProvider) Capabilities() types.ProviderCapabilities {
	return p.Caps
}

// Operations returns a copy of the provider-published operations
func (p *BaseProvider) Operations() []types.OperationDescriptor {
	if len(p.Ops) == 0 {
		return nil
	}

	out := make([]types.OperationDescriptor, len(p.Ops))
	copy(out, p.Ops)

	return out
}

// ClientDescriptors returns a copy of the provider-published client descriptors
func (p *BaseProvider) ClientDescriptors() []types.ClientDescriptor {
	if len(p.Clients) == 0 {
		return nil
	}

	out := make([]types.ClientDescriptor, len(p.Clients))
	copy(out, p.Clients)

	return out
}
