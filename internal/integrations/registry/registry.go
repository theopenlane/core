package registry

import (
	"context"
	"maps"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/config"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/catalog"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

// Registry exposes loaded provider configs and runtime providers to callers
type Registry struct {
	configs        map[types.ProviderType]config.ProviderSpec
	providers      map[types.ProviderType]providers.Provider
	clients        map[types.ProviderType][]types.ClientDescriptor
	operations     map[types.ProviderType][]types.OperationDescriptor
	mappingCatalog *mappingCatalog
}

// NewRegistry loads embedded provider specs and builds the registry using the catalog builders
func NewRegistry(ctx context.Context, overrides map[string]config.ProviderSpec) (*Registry, error) {
	loader := config.NewFSLoader(config.ProvidersFS, "providers")
	specs, err := loader.Load()
	if err != nil {
		return nil, err
	}

	if len(overrides) > 0 {
		specs = config.MergeProviderSpecs(ctx, specs, overrides)
	}

	builders := catalog.Builders()

	instance := &Registry{
		configs:        make(map[types.ProviderType]config.ProviderSpec, len(specs)),
		providers:      map[types.ProviderType]providers.Provider{},
		clients:        map[types.ProviderType][]types.ClientDescriptor{},
		operations:     map[types.ProviderType][]types.OperationDescriptor{},
		mappingCatalog: newMappingCatalog(),
	}

	maps.Copy(instance.configs, specs)

	builderIndex := lo.SliceToMap(builders, func(b providers.Builder) (types.ProviderType, providers.Builder) {
		return b.Type(), b
	})

	for providerType, spec := range instance.configs {
		builder, ok := builderIndex[providerType]
		if !ok {
			continue
		}

		provider, err := builder.Build(ctx, spec)
		if err != nil || provider == nil {
			logx.FromContext(ctx).Warn().Err(err).Str("provider", string(providerType)).Msg("provider build failed, marking inactive")
			instance.markProviderInactive(providerType, spec)

			continue
		}

		instance.applyProviderArtifacts(providerType, provider)
	}
	instance.rebuildMappingCatalog()

	return instance, nil
}

// markProviderInactive updates one provider spec to inactive in the registry config map.
func (r *Registry) markProviderInactive(providerType types.ProviderType, spec config.ProviderSpec) {
	spec.Active = lo.ToPtr(false)
	r.configs[providerType] = spec
}

// applyProviderArtifacts stores one provider instance and refreshes derived client and operation descriptors.
func (r *Registry) applyProviderArtifacts(providerType types.ProviderType, provider providers.Provider) {
	r.providers[providerType] = provider

	if clientProvider, ok := provider.(types.ClientProvider); ok {
		if descriptors := providerkit.SanitizeClientDescriptors(providerType, clientProvider.ClientDescriptors()); len(descriptors) > 0 {
			r.clients[providerType] = descriptors
		} else {
			delete(r.clients, providerType)
		}
	} else {
		delete(r.clients, providerType)
	}

	if operationProvider, ok := provider.(types.OperationProvider); ok {
		if descriptors := providerkit.SanitizeOperationDescriptors(providerType, operationProvider.Operations()); len(descriptors) > 0 {
			r.operations[providerType] = descriptors
		} else {
			delete(r.operations, providerType)
		}
	} else {
		delete(r.operations, providerType)
	}
}

// rebuildMappingCatalog reconstructs provider default mapping registrations from the current provider map.
func (r *Registry) rebuildMappingCatalog() {
	r.mappingCatalog = newMappingCatalog()

	for providerType, provider := range r.providers {
		mappingProvider, ok := provider.(types.MappingProvider)
		if !ok {
			continue
		}

		r.mappingCatalog.registerProvider(providerType, mappingProvider.DefaultMappings())
	}
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

// ProviderMetadataCatalog returns a copy of all provider metadata entries.
func (r *Registry) ProviderMetadataCatalog() map[types.ProviderType]types.ProviderConfig {
	return lo.MapEntries(r.configs, func(key types.ProviderType, spec config.ProviderSpec) (types.ProviderType, types.ProviderConfig) {
		return key, spec.ToProviderConfig()
	})
}

// ClientDescriptorCatalog returns a copy of all provider client descriptors.
func (r *Registry) ClientDescriptorCatalog() map[types.ProviderType][]types.ClientDescriptor {
	return lo.MapEntries(r.clients, func(provider types.ProviderType, descriptors []types.ClientDescriptor) (types.ProviderType, []types.ClientDescriptor) {
		copied := make([]types.ClientDescriptor, len(descriptors))
		copy(copied, descriptors)
		return provider, copied
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
	return lo.MapEntries(r.operations, func(provider types.ProviderType, descriptors []types.OperationDescriptor) (types.ProviderType, []types.OperationDescriptor) {
		copied := make([]types.OperationDescriptor, len(descriptors))
		copy(copied, descriptors)
		return provider, copied
	})
}

// MintCredential calls the registered provider's Mint method without accessing the credential store.
func (r *Registry) MintCredential(ctx context.Context, request types.CredentialMintRequest) (models.CredentialSet, error) {
	provider, ok := r.providers[request.Provider]
	if !ok {
		return models.CredentialSet{}, ErrProviderNotFound
	}

	return provider.Mint(ctx, request)
}

// UpsertProvider adds or replaces a provider/spec after initialization (primarily for tests).
func (r *Registry) UpsertProvider(ctx context.Context, spec config.ProviderSpec, builder providers.Builder) error {
	providerType := spec.ProviderType()
	if providerType == types.ProviderUnknown {
		return ErrProviderTypeRequired
	}
	if builder == nil || builder.Type() != providerType {
		return ErrBuilderMismatch
	}

	r.configs[providerType] = spec

	provider, err := builder.Build(ctx, spec)
	if err != nil {
		return ErrProviderBuildFailed
	}
	if provider == nil {
		return ErrProviderNil
	}

	r.applyProviderArtifacts(providerType, provider)
	r.rebuildMappingCatalog()

	return nil
}

// SupportsIngest reports whether a provider has any registered mappings for the schema.
func (r *Registry) SupportsIngest(provider types.ProviderType, schema types.MappingSchema) bool {
	return r.mappingCatalog.supports(provider, schema)
}

// DefaultMapping returns the mapping spec for a provider, schema, and variant.
// It checks for an exact variant match first, then falls back to the empty-variant default.
func (r *Registry) DefaultMapping(provider types.ProviderType, schema types.MappingSchema, variant string) (types.MappingSpec, bool) {
	return r.mappingCatalog.resolve(provider, schema, variant)
}
