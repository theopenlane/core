package registry

import (
	"context"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/spec"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

// Registry exposes loaded provider specs and runtime providers to callers
type Registry struct {
	configs        map[types.ProviderType]spec.ProviderSpec
	providers      map[types.ProviderType]types.Provider
	clients        map[types.ProviderType][]types.ClientDescriptor
	operations     map[types.ProviderType][]types.OperationDescriptor
	mappingCatalog *mappingCatalog
}

// NewRegistry builds the provider registry from the supplied builders.
// Each builder contributes its static spec via Spec() and a runtime instance via Build().
func NewRegistry(ctx context.Context, builders []providers.Builder) (*Registry, error) {
	instance := &Registry{
		configs:        make(map[types.ProviderType]spec.ProviderSpec, len(builders)),
		providers:      map[types.ProviderType]types.Provider{},
		clients:        map[types.ProviderType][]types.ClientDescriptor{},
		operations:     map[types.ProviderType][]types.OperationDescriptor{},
		mappingCatalog: newMappingCatalog(),
	}

	for _, builder := range builders {
		providerSpec := builder.Spec()
		providerType := providerSpec.ProviderType()
		instance.configs[providerType] = providerSpec

		provider, buildErr := builder.Build(ctx, providerSpec)
		if buildErr != nil || provider == nil {
			logx.FromContext(ctx).Warn().Err(buildErr).Str("provider", string(providerType)).Msg("provider build failed, marking inactive")
			instance.markProviderInactive(providerType, providerSpec)

			continue
		}

		instance.applyProviderArtifacts(providerType, provider)
	}

	instance.rebuildMappingCatalog()

	return instance, nil
}

// markProviderInactive updates one provider spec to inactive in the registry config map
func (r *Registry) markProviderInactive(providerType types.ProviderType, providerSpec spec.ProviderSpec) {
	providerSpec.Active = lo.ToPtr(false)
	r.configs[providerType] = providerSpec
}

// applyProviderArtifacts stores one provider instance and refreshes derived client and operation descriptors
func (r *Registry) applyProviderArtifacts(providerType types.ProviderType, provider types.Provider) {
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

// rebuildMappingCatalog reconstructs provider default mapping registrations from the current provider map
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
func (r *Registry) Config(provider types.ProviderType) (spec.ProviderSpec, bool) {
	providerSpec, ok := r.configs[provider]

	return providerSpec, ok
}

// ProviderMetadataCatalog returns a copy of all provider metadata entries
func (r *Registry) ProviderMetadataCatalog() map[types.ProviderType]types.IntegrationProviderMetadata {
	return lo.MapEntries(r.configs, func(key types.ProviderType, providerSpec spec.ProviderSpec) (types.ProviderType, types.IntegrationProviderMetadata) {
		return key, providerSpec.ToProviderMetadata()
	})
}

// ClientDescriptorCatalog returns a copy of all provider client descriptors
func (r *Registry) ClientDescriptorCatalog() map[types.ProviderType][]types.ClientDescriptor {
	return lo.MapEntries(r.clients, func(provider types.ProviderType, descriptors []types.ClientDescriptor) (types.ProviderType, []types.ClientDescriptor) {
		copied := make([]types.ClientDescriptor, len(descriptors))
		copy(copied, descriptors)

		return provider, copied
	})
}

// OperationDescriptors returns the registered operation descriptors for a provider
func (r *Registry) OperationDescriptors(provider types.ProviderType) []types.OperationDescriptor {
	descriptors := r.operations[provider]
	if len(descriptors) == 0 {
		return nil
	}

	out := make([]types.OperationDescriptor, len(descriptors))
	copy(out, descriptors)

	return out
}

// OperationDescriptorCatalog returns a copy of all provider operation descriptors
func (r *Registry) OperationDescriptorCatalog() map[types.ProviderType][]types.OperationDescriptor {
	return lo.MapEntries(r.operations, func(provider types.ProviderType, descriptors []types.OperationDescriptor) (types.ProviderType, []types.OperationDescriptor) {
		copied := make([]types.OperationDescriptor, len(descriptors))
		copy(copied, descriptors)

		return provider, copied
	})
}

// MintCredential calls the registered provider's Mint method without accessing the credential store
func (r *Registry) MintCredential(ctx context.Context, request types.CredentialMintRequest) (types.CredentialSet, error) {
	provider, ok := r.providers[request.Provider]
	if !ok {
		return types.CredentialSet{}, ErrProviderNotFound
	}

	return provider.Mint(ctx, request)
}

// UpsertProvider adds or replaces a provider/spec after initialization (primarily for tests)
func (r *Registry) UpsertProvider(ctx context.Context, providerSpec spec.ProviderSpec, builder providers.Builder) error {
	providerType := providerSpec.ProviderType()
	if providerType == types.ProviderUnknown {
		return ErrProviderTypeRequired
	}

	if builder == nil || builder.Type() != providerType {
		return ErrBuilderMismatch
	}

	r.configs[providerType] = providerSpec

	provider, err := builder.Build(ctx, providerSpec)
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

// DefaultMapping returns the mapping override for a provider, schema, and variant.
// It checks for an exact variant match first, then falls back to the empty-variant default.
func (r *Registry) DefaultMapping(provider types.ProviderType, schema types.MappingSchema, variant string) (types.MappingOverride, bool) {
	return r.mappingCatalog.resolve(provider, schema, variant)
}
