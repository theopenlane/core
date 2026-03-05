package ingest

import (
	"github.com/samber/lo"

	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	integrationtypes "github.com/theopenlane/core/internal/integrations/types"
)

var generatedMappingSchemas = lo.SliceToMap(lo.Keys(integrationgenerated.IntegrationMappingSchemas), func(schemaName string) (string, integrationtypes.MappingSchema) {
	return schemaName, integrationtypes.MappingSchema(schemaName)
})

// defaultMappingSpec resolves the built-in mapping override for a provider, schema, and variant
func defaultMappingSpec(mappingIndex integrationtypes.MappingIndex, provider integrationtypes.ProviderType, schemaName string, variant string) (openapi.IntegrationMappingOverride, bool) {
	if mappingIndex == nil {
		return openapi.IntegrationMappingOverride{}, false
	}

	schema, ok := schemaNameToMappingSchema(schemaName)
	if !ok {
		return openapi.IntegrationMappingOverride{}, false
	}

	spec, ok := mappingIndex.DefaultMapping(provider, schema, variant)
	if !ok {
		return openapi.IntegrationMappingOverride{}, false
	}

	return openapi.IntegrationMappingOverride{FilterExpr: spec.FilterExpr, MapExpr: spec.MapExpr}, true
}

// supportsDefaultMapping reports whether the provider has built-in mappings registered for the given schema.
func supportsDefaultMapping(mappingIndex integrationtypes.MappingIndex, provider integrationtypes.ProviderType, schemaName string) bool {
	if mappingIndex == nil {
		return false
	}

	schema, ok := schemaNameToMappingSchema(schemaName)
	if !ok {
		return false
	}

	return mappingIndex.SupportsIngest(provider, schema)
}

// supportsSchemaIngest reports whether ingest mappings are available for one provider/schema
// from either provider defaults or integration mapping overrides.
func supportsSchemaIngest(mappingIndex integrationtypes.MappingIndex, provider integrationtypes.ProviderType, integrationConfig openapi.IntegrationConfig, schema integrationtypes.MappingSchema) bool {
	schemaName := string(integrationtypes.NormalizeMappingSchema(schema))
	if supportsDefaultMapping(mappingIndex, provider, schemaName) {
		return true
	}

	return newMappingOverrideIndex(integrationConfig).HasAny(provider, schemaName)
}

// schemaNameToMappingSchema resolves a generated schema name to the shared mapping schema type.
func schemaNameToMappingSchema(schemaName string) (integrationtypes.MappingSchema, bool) {
	schema, ok := generatedMappingSchemas[schemaName]

	return schema, ok
}
