package registry

import (
	"context"
	"fmt"
	"maps"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/config"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/catalog"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Registry exposes loaded provider configs and runtime providers to callers
type Registry struct {
	configs    map[types.ProviderType]config.ProviderSpec
	providers  map[types.ProviderType]providers.Provider
	clients    map[types.ProviderType][]types.ClientDescriptor
	operations map[types.ProviderType][]types.OperationDescriptor
}

// NewRegistry builds a registry from the supplied specs and factories
func NewRegistry(ctx context.Context, specs map[types.ProviderType]config.ProviderSpec, builders []providers.Builder) (*Registry, error) {
	if len(specs) == 0 {
		return nil, ErrNoProviderSpecs
	}

	instance := &Registry{
		configs:    specs,
		providers:  map[types.ProviderType]providers.Provider{},
		clients:    map[types.ProviderType][]types.ClientDescriptor{},
		operations: map[types.ProviderType][]types.OperationDescriptor{},
	}

	builderIndex := lo.SliceToMap(builders, func(b providers.Builder) (types.ProviderType, providers.Builder) {
		return b.Type(), b
	})

	for providerType, spec := range specs {
		builder, ok := builderIndex[providerType]
		if !ok {
			continue
		}

		provider, err := builder.Build(ctx, spec)
		if err != nil {
			return nil, fmt.Errorf("integrations/registry: build provider %s: %w", providerType, err)
		}

		instance.providers[providerType] = provider

		if clientProvider, ok := provider.(types.ClientProvider); ok {
			if descriptors := sanitizeDescriptors(providerType, clientProvider.ClientDescriptors()); len(descriptors) > 0 {
				instance.clients[providerType] = descriptors
			}
		}

		if operationProvider, ok := provider.(types.OperationProvider); ok {
			if ops := sanitizeOperationDescriptors(providerType, operationProvider.Operations()); len(ops) > 0 {
				instance.operations[providerType] = ops
			}
		}
	}

	return instance, nil
}

// Provider returns a registered provider instance
func (r *Registry) Provider(provider types.ProviderType) (types.Provider, bool) {
	value, ok := r.providers[provider]

	return value, ok
}

// Config returns the raw provider specification for declarative handlers
func (r *Registry) Config(provider types.ProviderType) (config.ProviderSpec, bool) {
	spec, ok := r.configs[provider]

	return spec, ok
}

// ProviderConfigs exposes the full provider config map (copy) for consumers needing iteration
func (r *Registry) ProviderConfigs() map[types.ProviderType]config.ProviderSpec {
	out := make(map[types.ProviderType]config.ProviderSpec, len(r.configs))
	maps.Copy(out, r.configs)

	return out
}

// ProviderMetadata returns the handler-facing provider metadata (docs, schema, etc.).
func (r *Registry) ProviderMetadata(provider types.ProviderType) (types.ProviderConfig, bool) {
	spec, ok := r.configs[provider]
	if !ok {
		return types.ProviderConfig{}, false
	}

	return spec.ToProviderConfig(), true
}

// ProviderMetadataCatalog returns a copy of all provider metadata entries.
func (r *Registry) ProviderMetadataCatalog() map[types.ProviderType]types.ProviderConfig {
	out := make(map[types.ProviderType]types.ProviderConfig, len(r.configs))

	for key, spec := range r.configs {
		out[key] = spec.ToProviderConfig()
	}

	return out
}

// ClientDescriptors returns the registered client descriptors for a provider.
func (r *Registry) ClientDescriptors(provider types.ProviderType) []types.ClientDescriptor {
	descriptors := r.clients[provider]
	if len(descriptors) == 0 {
		return nil
	}

	out := make([]types.ClientDescriptor, len(descriptors))
	copy(out, descriptors)

	return out
}

// ClientDescriptorCatalog returns a copy of all provider client descriptors.
func (r *Registry) ClientDescriptorCatalog() map[types.ProviderType][]types.ClientDescriptor {
	out := make(map[types.ProviderType][]types.ClientDescriptor, len(r.clients))
	for provider, descriptors := range r.clients {
		copied := make([]types.ClientDescriptor, len(descriptors))

		copy(copied, descriptors)

		out[provider] = copied
	}

	return out
}

// OperationDescriptors returns the registered operation descriptors for a provider.
func (r *Registry) OperationDescriptors(provider types.ProviderType) []types.OperationDescriptor {
	descriptors := r.operations[provider]
	if len(descriptors) == 0 {
		return nil
	}

	out := make([]types.OperationDescriptor, len(descriptors))
	copy(out, descriptors)

	return out
}

// OperationDescriptorCatalog returns a copy of all provider operation descriptors.
func (r *Registry) OperationDescriptorCatalog() map[types.ProviderType][]types.OperationDescriptor {
	out := make(map[types.ProviderType][]types.OperationDescriptor, len(r.operations))

	for provider, descriptors := range r.operations {
		copied := make([]types.OperationDescriptor, len(descriptors))
		copy(copied, descriptors)
		out[provider] = copied
	}

	return out
}

// LoadRegistry loads provider specs using the supplied loader and builder catalog.
func LoadRegistry(ctx context.Context, loader *config.FSLoader) (*Registry, error) {
	specs, err := loader.Load(ctx)
	if err != nil {
		return nil, err
	}

	builders := catalog.Builders()

	return NewRegistry(ctx, specs, builders)
}

// LoadDefaultRegistry loads provider specs from the embedded filesystem.
func LoadDefaultRegistry(ctx context.Context) (*Registry, error) {
	loader := config.NewFSLoader(config.ProvidersFS, "providers")

	return LoadRegistry(ctx, loader)
}

// sanitizeDescriptors filters out invalid client descriptors and assigns provider type
func sanitizeDescriptors(provider types.ProviderType, descriptors []types.ClientDescriptor) []types.ClientDescriptor {
	if len(descriptors) == 0 {
		return nil
	}

	out := make([]types.ClientDescriptor, 0, len(descriptors))
	for _, descriptor := range descriptors {
		if descriptor.Build == nil {
			continue
		}
		if descriptor.Provider == types.ProviderUnknown {
			descriptor.Provider = provider
		}
		out = append(out, descriptor)
	}

	return out
}

// sanitizeOperationDescriptors filters out invalid operation descriptors and assigns provider type
func sanitizeOperationDescriptors(provider types.ProviderType, descriptors []types.OperationDescriptor) []types.OperationDescriptor {
	if len(descriptors) == 0 {
		return nil
	}

	out := make([]types.OperationDescriptor, 0, len(descriptors))

	for _, descriptor := range descriptors {
		if descriptor.Run == nil {
			continue
		}

		if descriptor.Name == "" {
			continue
		}

		if descriptor.Provider == types.ProviderUnknown {
			descriptor.Provider = provider
		}

		out = append(out, descriptor)
	}

	return out
}
