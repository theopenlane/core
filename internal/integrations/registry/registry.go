package registry

import (
	"context"
	"fmt"
	"maps"

	"github.com/samber/lo"

	commonhelpers "github.com/theopenlane/core/common/helpers"
	"github.com/theopenlane/core/common/integrations/config"
	"github.com/theopenlane/core/common/integrations/helpers"
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/catalog"
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
			return nil, fmt.Errorf("%w: %w", ErrProviderBuildFailed, err)
		}

		if provider == nil {
			return nil, fmt.Errorf("%w: provider is nil", ErrProviderBuildFailed)
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
	return commonhelpers.FoldMap(r.configs, func(capacity int) map[types.ProviderType]types.ProviderConfig {
		return make(map[types.ProviderType]types.ProviderConfig, capacity)
	}, func(acc *map[types.ProviderType]types.ProviderConfig, key types.ProviderType, spec config.ProviderSpec) {
		(*acc)[key] = spec.ToProviderConfig()
	})
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
	return commonhelpers.FoldMap(r.clients, func(capacity int) map[types.ProviderType][]types.ClientDescriptor {
		return make(map[types.ProviderType][]types.ClientDescriptor, capacity)
	}, func(acc *map[types.ProviderType][]types.ClientDescriptor, provider types.ProviderType, descriptors []types.ClientDescriptor) {
		copied := make([]types.ClientDescriptor, len(descriptors))
		copy(copied, descriptors)
		(*acc)[provider] = copied
	})
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
	return commonhelpers.FoldMap(r.operations, func(capacity int) map[types.ProviderType][]types.OperationDescriptor {
		return make(map[types.ProviderType][]types.OperationDescriptor, capacity)
	}, func(acc *map[types.ProviderType][]types.OperationDescriptor, provider types.ProviderType, descriptors []types.OperationDescriptor) {
		copied := make([]types.OperationDescriptor, len(descriptors))
		copy(copied, descriptors)
		(*acc)[provider] = copied
	})
}

// UpsertProvider adds or replaces a provider/spec after initialization (primarily for tests).
func (r *Registry) UpsertProvider(ctx context.Context, spec config.ProviderSpec, builder providers.Builder) error {
	if r == nil {
		return ErrRegistryNil
	}

	providerType := spec.ProviderType()
	if providerType == types.ProviderUnknown {
		return ErrProviderTypeRequired
	}
	if builder == nil || builder.Type() != providerType {
		return ErrBuilderMismatch
	}

	provider, err := builder.Build(ctx, spec)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProviderBuildFailed, err)
	}
	if provider == nil {
		return fmt.Errorf("%w: provider is nil", ErrProviderBuildFailed)
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
