package ingest

import (
	"github.com/samber/lo"
	integrationtypes "github.com/theopenlane/core/common/integrations/types"
	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
)

var normalizedGeneratedMappingSchemas = lo.SliceToMap(lo.Keys(integrationgenerated.IntegrationMappingSchemas), func(schemaName string) (string, integrationtypes.MappingSchema) {
	return normalizeMappingKey(schemaName), integrationtypes.MappingSchema(schemaName)
})

// mappingIndexSource is the package-level MappingIndex used to resolve provider default mappings.
// It is set at server startup via SetMappingIndex.
var mappingIndexSource integrationtypes.MappingIndex

// SetMappingIndex registers the MappingIndex used to resolve provider-registered default mappings.
// It must be called before any ingest functions that rely on default mappings are invoked.
func SetMappingIndex(index integrationtypes.MappingIndex) {
	mappingIndexSource = index
}

// defaultMappingSpec resolves the built-in mapping override for a provider, schema, and variant.
func defaultMappingSpec(provider integrationtypes.ProviderType, schemaName string, variant string) (openapi.IntegrationMappingOverride, bool) {
	if mappingIndexSource == nil {
		return openapi.IntegrationMappingOverride{}, false
	}

	schema, ok := schemaNameToMappingSchema(schemaName)
	if !ok {
		return openapi.IntegrationMappingOverride{}, false
	}

	spec, ok := mappingIndexSource.DefaultMapping(provider, schema, variant)
	if !ok {
		return openapi.IntegrationMappingOverride{}, false
	}

	return openapi.IntegrationMappingOverride{FilterExpr: spec.FilterExpr, MapExpr: spec.MapExpr}, true
}

// supportsDefaultMapping reports whether the provider has built-in mappings registered for the given schema.
func supportsDefaultMapping(provider integrationtypes.ProviderType, schemaName string) bool {
	if mappingIndexSource == nil {
		return false
	}

	schema, ok := schemaNameToMappingSchema(schemaName)
	if !ok {
		return false
	}

	return mappingIndexSource.SupportsIngest(provider, schema)
}

func schemaNameToMappingSchema(schemaName string) (integrationtypes.MappingSchema, bool) {
	schema, ok := normalizedGeneratedMappingSchemas[normalizeMappingKey(schemaName)]

	return schema, ok
}
