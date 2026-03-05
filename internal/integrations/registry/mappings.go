package registry

import (
	"strings"

	"github.com/theopenlane/core/internal/integrations/types"
)

// mappingKey uniquely identifies a provider mapping by provider, schema, and variant
type mappingKey struct {
	provider types.ProviderType
	schema   types.MappingSchema
	variant  string
}

// providerSchemaKey identifies whether any mapping exists for a provider and schema pair
type providerSchemaKey struct {
	provider types.ProviderType
	schema   types.MappingSchema
}

// mappingCatalog stores provider default mappings in a typed lookup index
type mappingCatalog struct {
	byKey              map[mappingKey]types.MappingSpec
	supportsByProvider map[providerSchemaKey]struct{}
}

// newMappingCatalog builds an empty provider mapping catalog
func newMappingCatalog() *mappingCatalog {
	return &mappingCatalog{
		byKey:              map[mappingKey]types.MappingSpec{},
		supportsByProvider: map[providerSchemaKey]struct{}{},
	}
}

// register records one provider mapping in the catalog
func (c *mappingCatalog) register(provider types.ProviderType, schema types.MappingSchema, variant string, spec types.MappingSpec) {
	normalizedSchema := types.NormalizeMappingSchema(schema)
	if provider == types.ProviderUnknown || normalizedSchema == "" {
		return
	}

	key := mappingKey{
		provider: provider,
		schema:   normalizedSchema,
		variant:  strings.TrimSpace(variant),
	}
	c.byKey[key] = spec
	c.supportsByProvider[providerSchemaKey{
		provider: provider,
		schema:   normalizedSchema,
	}] = struct{}{}
}

// registerProvider records all mappings published by one provider
func (c *mappingCatalog) registerProvider(provider types.ProviderType, mappings []types.MappingRegistration) {
	if provider == types.ProviderUnknown {
		return
	}

	for _, mapping := range mappings {
		c.register(provider, mapping.Schema, mapping.Variant, mapping.Spec)
	}
}

// supports reports whether any mapping exists for provider and schema
func (c *mappingCatalog) supports(provider types.ProviderType, schema types.MappingSchema) bool {
	_, ok := c.supportsByProvider[providerSchemaKey{
		provider: provider,
		schema:   types.NormalizeMappingSchema(schema),
	}]

	return ok
}

// resolve returns a mapping for provider/schema/variant, falling back to empty-variant default
func (c *mappingCatalog) resolve(provider types.ProviderType, schema types.MappingSchema, variant string) (types.MappingSpec, bool) {
	normalizedSchema := types.NormalizeMappingSchema(schema)
	if normalizedSchema == "" || provider == types.ProviderUnknown {
		return types.MappingSpec{}, false
	}

	normalizedVariant := strings.TrimSpace(variant)
	if normalizedVariant != "" {
		if spec, ok := c.byKey[mappingKey{
			provider: provider,
			schema:   normalizedSchema,
			variant:  normalizedVariant,
		}]; ok {
			return spec, true
		}
	}

	spec, ok := c.byKey[mappingKey{
		provider: provider,
		schema:   normalizedSchema,
		variant:  "",
	}]

	return spec, ok
}
