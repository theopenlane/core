package registry

import (
	"context"
	"fmt"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/config"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Registry exposes loaded provider configs and runtime providers to callers
type Registry struct {
	configs   map[types.ProviderType]config.ProviderSpec
	providers map[types.ProviderType]providers.Provider
}

// New builds a registry from the supplied specs and factories
func New(ctx context.Context, specs map[types.ProviderType]config.ProviderSpec, builders []providers.Builder) (*Registry, error) {
	if len(specs) == 0 {
		return nil, fmt.Errorf("integrations/registry: no provider specs supplied")
	}

	instance := &Registry{
		configs:   specs,
		providers: map[types.ProviderType]providers.Provider{},
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
	}

	return instance, nil
}

// Provider returns a registered provider instance
func (r *Registry) Provider(provider types.ProviderType) (types.Provider, bool) {
	if r == nil {
		return nil, false
	}

	value, ok := r.providers[provider]

	return value, ok
}

// Config returns the raw provider specification for declarative handlers
func (r *Registry) Config(provider types.ProviderType) (config.ProviderSpec, bool) {
	if r == nil {
		return config.ProviderSpec{}, false
	}

	spec, ok := r.configs[provider]

	return spec, ok
}

// ProviderConfigs exposes the full provider config map (copy) for consumers needing iteration
func (r *Registry) ProviderConfigs() map[types.ProviderType]config.ProviderSpec {
	if r == nil {
		return nil
	}

	out := make(map[types.ProviderType]config.ProviderSpec, len(r.configs))
	for key, value := range r.configs {
		out[key] = value
	}

	return out
}

// ProviderMetadata returns the handler-facing provider metadata (docs, schema, etc.).
func (r *Registry) ProviderMetadata(provider types.ProviderType) (types.ProviderConfig, bool) {
	if r == nil {
		return types.ProviderConfig{}, false
	}

	spec, ok := r.configs[provider]
	if !ok {
		return types.ProviderConfig{}, false
	}

	return spec.ToProviderConfig(), true
}

// ProviderMetadataCatalog returns a copy of all provider metadata entries.
func (r *Registry) ProviderMetadataCatalog() map[types.ProviderType]types.ProviderConfig {
	if r == nil {
		return nil
	}

	out := make(map[types.ProviderType]types.ProviderConfig, len(r.configs))
	for key, spec := range r.configs {
		out[key] = spec.ToProviderConfig()
	}

	return out
}
