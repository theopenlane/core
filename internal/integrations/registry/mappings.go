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

// mappingCatalog stores provider default mappings in a typed lookup index
type mappingCatalog struct {
	byKey map[mappingKey]types.MappingOverride
}

// newMappingCatalog builds an empty provider mapping catalog
func newMappingCatalog() *mappingCatalog {
	return &mappingCatalog{
		byKey: map[mappingKey]types.MappingOverride{},
	}
}

// register records one provider mapping in the catalog
func (c *mappingCatalog) register(provider types.ProviderType, schema types.MappingSchema, variant string, override types.MappingOverride) {
	normalizedSchema := normalizeMappingSchema(schema)
	if provider == types.ProviderUnknown || normalizedSchema == "" {
		return
	}

	key := mappingKey{
		provider: provider,
		schema:   normalizedSchema,
		variant:  strings.TrimSpace(variant),
	}

	c.byKey[key] = override
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

// resolve returns a mapping override for provider/schema/variant, falling back to empty-variant default
func (c *mappingCatalog) resolve(provider types.ProviderType, schema types.MappingSchema, variant string) (types.MappingOverride, bool) {
	normalizedSchema := normalizeMappingSchema(schema)
	if normalizedSchema == "" || provider == types.ProviderUnknown {
		return types.MappingOverride{}, false
	}

	normalizedVariant := strings.TrimSpace(variant)
	if normalizedVariant != "" {
		if override, ok := c.byKey[mappingKey{
			provider: provider,
			schema:   normalizedSchema,
			variant:  normalizedVariant,
		}]; ok {
			return override, true
		}
	}

	override, ok := c.byKey[mappingKey{
		provider: provider,
		schema:   normalizedSchema,
		variant:  "",
	}]

	return override, ok
}

// normalizeMappingSchema trims whitespace from a schema name and returns the result
func normalizeMappingSchema(schema types.MappingSchema) types.MappingSchema {
	value := strings.TrimSpace(string(schema))
	if value == "" {
		return ""
	}

	return types.MappingSchema(value)
}
