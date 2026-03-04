package types

import "strings"

// MappingSpec holds a provider-defined filter and mapping expression pair used during ingest.
type MappingSpec struct {
	// FilterExpr is a CEL expression evaluated against the raw provider payload; return true to include the record.
	FilterExpr string
	// MapExpr is a CEL expression that projects the raw payload into the normalized output schema.
	MapExpr string
}

// MappingSchema identifies a normalized ingest schema (for example Vulnerability).
type MappingSchema string

const (
	// MappingSchemaVulnerability identifies vulnerability ingest mappings.
	MappingSchemaVulnerability MappingSchema = "Vulnerability"
	// MappingSchemaDirectoryAccount identifies directory account ingest mappings.
	MappingSchemaDirectoryAccount MappingSchema = "DirectoryAccount"
)

// NormalizeMappingSchema trims and canonicalizes schema names.
func NormalizeMappingSchema(schema MappingSchema) MappingSchema {
	value := strings.TrimSpace(string(schema))
	if value == "" {
		return ""
	}

	return MappingSchema(value)
}

// MappingRegistration declares a provider mapping for one schema and variant.
type MappingRegistration struct {
	// Schema identifies the normalized ingest schema.
	Schema MappingSchema
	// Variant scopes the mapping within the schema (empty string = default).
	Variant string
	// Spec contains CEL filter/map expressions.
	Spec MappingSpec
}

// MappingProvider is implemented by providers that supply built-in ingest mappings.
type MappingProvider interface {
	Provider
	// DefaultMappings returns built-in mappings across one or more schemas.
	DefaultMappings() []MappingRegistration
}

// MappingKey uniquely identifies a provider mapping by provider, schema, and variant.
type MappingKey struct {
	Provider ProviderType
	Schema   MappingSchema
	Variant  string
}

type providerSchemaKey struct {
	Provider ProviderType
	Schema   MappingSchema
}

// MappingCatalog stores provider default mappings in a typed lookup index.
type MappingCatalog struct {
	byKey              map[MappingKey]MappingSpec
	supportsByProvider map[providerSchemaKey]struct{}
}

// NewMappingCatalog builds an empty provider mapping catalog.
func NewMappingCatalog() *MappingCatalog {
	return &MappingCatalog{
		byKey:              map[MappingKey]MappingSpec{},
		supportsByProvider: map[providerSchemaKey]struct{}{},
	}
}

// Register records one provider mapping in the catalog.
func (c *MappingCatalog) Register(provider ProviderType, schema MappingSchema, variant string, spec MappingSpec) {
	normalizedSchema := NormalizeMappingSchema(schema)
	if provider == ProviderUnknown || normalizedSchema == "" {
		return
	}

	key := MappingKey{
		Provider: provider,
		Schema:   normalizedSchema,
		Variant:  strings.TrimSpace(variant),
	}
	c.byKey[key] = spec
	c.supportsByProvider[providerSchemaKey{
		Provider: provider,
		Schema:   normalizedSchema,
	}] = struct{}{}
}

// RegisterProvider records all mappings published by one provider.
func (c *MappingCatalog) RegisterProvider(provider ProviderType, mappings []MappingRegistration) {
	if provider == ProviderUnknown {
		return
	}

	for _, mapping := range mappings {
		c.Register(provider, mapping.Schema, mapping.Variant, mapping.Spec)
	}
}

// Supports reports whether any mapping exists for provider and schema.
func (c *MappingCatalog) Supports(provider ProviderType, schema MappingSchema) bool {
	_, ok := c.supportsByProvider[providerSchemaKey{
		Provider: provider,
		Schema:   NormalizeMappingSchema(schema),
	}]

	return ok
}

// Resolve returns a mapping for provider/schema/variant, falling back to empty-variant default.
func (c *MappingCatalog) Resolve(provider ProviderType, schema MappingSchema, variant string) (MappingSpec, bool) {
	normalizedSchema := NormalizeMappingSchema(schema)
	if normalizedSchema == "" || provider == ProviderUnknown {
		return MappingSpec{}, false
	}

	normalizedVariant := strings.TrimSpace(variant)
	if normalizedVariant != "" {
		if spec, ok := c.byKey[MappingKey{
			Provider: provider,
			Schema:   normalizedSchema,
			Variant:  normalizedVariant,
		}]; ok {
			return spec, true
		}
	}

	spec, ok := c.byKey[MappingKey{
		Provider: provider,
		Schema:   normalizedSchema,
		Variant:  "",
	}]

	return spec, ok
}

// MappingIndex resolves provider-registered default mapping specs by schema and variant.
// It is implemented by the integration registry and injected during server startup.
type MappingIndex interface {
	// SupportsIngest reports whether the provider has any registered mappings for the schema.
	SupportsIngest(provider ProviderType, schema MappingSchema) bool
	// DefaultMapping returns the mapping spec for provider/schema/variant.
	DefaultMapping(provider ProviderType, schema MappingSchema, variant string) (MappingSpec, bool)
}
