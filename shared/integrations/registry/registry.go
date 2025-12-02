package registry

import (
	"context"
	"errors"
	"fmt"
	"maps"

	"github.com/samber/lo"

	"github.com/theopenlane/shared/integrations/config"
	"github.com/theopenlane/shared/integrations/providers"
	"github.com/theopenlane/shared/integrations/providers/catalog"
	"github.com/theopenlane/shared/integrations/providers/helpers"
	"github.com/theopenlane/shared/integrations/types"
)

var (
	errRegistryNil     = errors.New("integrations/registry: registry is nil")
	errProviderType    = errors.New("integrations/registry: provider type required")
	errBuilderMismatch = errors.New("integrations/registry: builder missing or type mismatch")
)

// Registry exposes loaded provider configs and runtime providers to callers
type Registry struct {
	configs    map[types.ProviderType]config.ProviderSpec
	providers  map[types.ProviderType]providers.Provider
	clients    map[types.ProviderType][]types.ClientDescriptor
	operations map[types.ProviderType][]types.OperationDescriptor
}

// NewRegistry loads embedded provider specs and builds the registry using the catalog builders.
func NewRegistry(ctx context.Context) (*Registry, error) {
	loader := config.NewFSLoader(config.ProvidersFS, "providers")
	specs, err := loader.Load()
	if err != nil {
		return nil, err
	}

	builders := catalog.Builders()

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
			if descriptors := helpers.SanitizeClientDescriptors(providerType, clientProvider.ClientDescriptors()); len(descriptors) > 0 {
				instance.clients[providerType] = descriptors
			}
		}

		if operationProvider, ok := provider.(types.OperationProvider); ok {
			if ops := helpers.SanitizeOperationDescriptors(providerType, operationProvider.Operations()); len(ops) > 0 {
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

// UpsertProvider adds or replaces a provider/spec after initialization (primarily for tests).
func (r *Registry) UpsertProvider(ctx context.Context, spec config.ProviderSpec, builder providers.Builder) error {
	if r == nil {
		return errRegistryNil
	}

	providerType := spec.ProviderType()
	if providerType == types.ProviderUnknown {
		return errProviderType
	}
	if builder == nil || builder.Type() != providerType {
		return fmt.Errorf("%w for %s", errBuilderMismatch, providerType)
	}

	provider, err := builder.Build(ctx, spec)
	if err != nil {
		return fmt.Errorf("integrations/registry: build provider %s: %w", providerType, err)
	}

	r.configs[providerType] = spec
	r.providers[providerType] = provider

	if clientProvider, ok := provider.(types.ClientProvider); ok {
		if descriptors := helpers.SanitizeClientDescriptors(providerType, clientProvider.ClientDescriptors()); len(descriptors) > 0 {
			r.clients[providerType] = descriptors
		} else {
			delete(r.clients, providerType)
		}
	} else {
		delete(r.clients, providerType)
	}

	if operationProvider, ok := provider.(types.OperationProvider); ok {
		if ops := helpers.SanitizeOperationDescriptors(providerType, operationProvider.Operations()); len(ops) > 0 {
			r.operations[providerType] = ops
		} else {
			delete(r.operations, providerType)
		}
	} else {
		delete(r.operations, providerType)
	}

	return nil
}
