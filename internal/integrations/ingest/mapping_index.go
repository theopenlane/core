package ingest

import (
	integrationtypes "github.com/theopenlane/core/common/integrations/types"
	openapi "github.com/theopenlane/core/common/openapi"
)

var (
	normalizedVulnerabilitySchema    = normalizeMappingKey(mappingSchemaVulnerability)
	normalizedDirectoryAccountSchema = normalizeMappingKey(mappingSchemaDirectoryAccount)
)

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

	switch normalizeMappingKey(schemaName) {
	case normalizedVulnerabilitySchema:
		spec, ok := mappingIndexSource.DefaultVulnerabilityMapping(provider, variant)
		if !ok {
			return openapi.IntegrationMappingOverride{}, false
		}

		return openapi.IntegrationMappingOverride{FilterExpr: spec.FilterExpr, MapExpr: spec.MapExpr}, true
	case normalizedDirectoryAccountSchema:
		spec, ok := mappingIndexSource.DefaultDirectoryAccountMapping(provider, variant)
		if !ok {
			return openapi.IntegrationMappingOverride{}, false
		}

		return openapi.IntegrationMappingOverride{FilterExpr: spec.FilterExpr, MapExpr: spec.MapExpr}, true
	default:
		return openapi.IntegrationMappingOverride{}, false
	}
}

// supportsDefaultMapping reports whether the provider has built-in mappings registered for the given schema.
func supportsDefaultMapping(provider integrationtypes.ProviderType, schemaName string) bool {
	if mappingIndexSource == nil {
		return false
	}

	switch normalizeMappingKey(schemaName) {
	case normalizedVulnerabilitySchema:
		return mappingIndexSource.SupportsVulnerabilityIngest(provider)
	case normalizedDirectoryAccountSchema:
		return mappingIndexSource.SupportsDirectoryAccountIngest(provider)
	default:
		return false
	}
}
