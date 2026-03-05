package types

import (
	"strings"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
)

// MappingSpec holds a provider-defined filter and mapping expression pair used during ingest
type MappingSpec struct {
	// FilterExpr is a CEL expression evaluated against the raw provider payload; return true to include the record.
	FilterExpr string
	// MapExpr is a CEL expression that projects the raw payload into the normalized output schema.
	MapExpr string
}

// MappingSchema identifies a normalized ingest schema (for example Vulnerability)
type MappingSchema string

const (
	// MappingSchemaVulnerability identifies vulnerability ingest mappings.
	MappingSchemaVulnerability MappingSchema = integrationgenerated.IntegrationMappingSchemaVulnerability
	// MappingSchemaDirectoryAccount identifies directory account ingest mappings.
	MappingSchemaDirectoryAccount MappingSchema = integrationgenerated.IntegrationMappingSchemaDirectoryAccount
)

// NormalizeMappingSchema trims and canonicalizes schema names
func NormalizeMappingSchema(schema MappingSchema) MappingSchema {
	value := strings.TrimSpace(string(schema))
	if value == "" {
		return ""
	}

	return MappingSchema(value)
}

// MappingRegistration declares a provider mapping for one schema and variant
type MappingRegistration struct {
	// Schema identifies the normalized ingest schema.
	Schema MappingSchema
	// Variant scopes the mapping within the schema (empty string = default).
	Variant string
	// Spec contains CEL filter/map expressions.
	Spec MappingSpec
}

// MappingProvider is implemented by providers that supply built-in ingest mappings
type MappingProvider interface {
	Provider
	// DefaultMappings returns built-in mappings across one or more schemas.
	DefaultMappings() []MappingRegistration
}

// MappingIndex resolves provider-registered default mapping specs by schema and variant;
// it is implemented by the integration registry and injected during server startup
type MappingIndex interface {
	// SupportsIngest reports whether the provider has any registered mappings for the schema.
	SupportsIngest(provider ProviderType, schema MappingSchema) bool
	// DefaultMapping returns the mapping spec for provider/schema/variant.
	DefaultMapping(provider ProviderType, schema MappingSchema, variant string) (MappingSpec, bool)
}
