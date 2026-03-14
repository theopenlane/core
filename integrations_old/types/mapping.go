package types

// MappingSchema identifies a normalized ingest schema (e.g. Vulnerability)
type MappingSchema string

const (
	// MappingSchemaVulnerability is the normalized schema name for vulnerability records
	MappingSchemaVulnerability MappingSchema = "Vulnerability"
	// MappingSchemaDirectoryAccount is the normalized schema name for directory account records
	MappingSchemaDirectoryAccount MappingSchema = "DirectoryAccount"
	// MappingSchemaDirectoryGroup is the normalized schema name for directory group records
	MappingSchemaDirectoryGroup MappingSchema = "DirectoryGroup"
	// MappingSchemaDirectoryMembership is the normalized schema name for directory membership records
	MappingSchemaDirectoryMembership MappingSchema = "DirectoryMembership"
	// MappingSchemaAsset is the normalized schema name for asset records
	MappingSchemaAsset MappingSchema = "Asset"
	// MappingSchemaContact is the normalized schema name for contact records
	MappingSchemaContact MappingSchema = "Contact"
	// MappingSchemaEntity is the normalized schema name for entity records
	MappingSchemaEntity MappingSchema = "Entity"
	// MappingSchemaRisk is the normalized schema name for risk records
	MappingSchemaRisk MappingSchema = "Risk"
)

// MappingOverride is a unified replacement for both types.MappingSpec and
// common/openapi.IntegrationMappingOverride; it holds user-configurable CEL
// expressions evaluated at ingest time via providerkit helpers —
// an empty string for FilterExpr or MapExpr is a no-op pass-through
type MappingOverride struct {
	// Version is the schema version for this override
	Version string `json:"version,omitempty"`
	// FilterExpr is a CEL expression evaluated against the raw provider payload;
	// return true to include the record, empty string is a no-op pass-through
	FilterExpr string `json:"filterExpr,omitempty"`
	// MapExpr is a CEL expression that projects the raw payload into the normalized
	// output schema, empty string is a no-op pass-through
	MapExpr string `json:"mapExpr,omitempty"`
}

// MappingRegistration declares a provider mapping for one schema and variant
type MappingRegistration struct {
	// Schema identifies the normalized ingest schema
	Schema MappingSchema
	// Variant scopes the mapping within the schema (empty string = default)
	Variant string
	// Spec contains CEL filter/map expressions
	Spec MappingOverride
}

// MappingIndex resolves provider-registered default mapping specs by schema and variant;
// it is implemented by the integration registry and injected during server startup
type MappingIndex interface {
	// DefaultMapping returns the mapping override for provider/schema/variant
	DefaultMapping(provider ProviderType, schema MappingSchema, variant string) (MappingOverride, bool)
}
