package types

// MappingSpec holds a provider-defined filter and mapping expression pair used during ingest.
type MappingSpec struct {
	// FilterExpr is a CEL expression evaluated against the raw provider payload; return true to include the record.
	FilterExpr string
	// MapExpr is a CEL expression that projects the raw payload into the normalized output schema.
	MapExpr string
}

// VulnerabilityMappingProvider is implemented by providers that supply built-in vulnerability ingest mappings.
type VulnerabilityMappingProvider interface {
	Provider
	// DefaultVulnerabilityMappings returns built-in mappings keyed by alert variant (e.g. "dependabot", "code_scanning").
	DefaultVulnerabilityMappings() map[string]MappingSpec
}

// DirectoryAccountMappingProvider is implemented by providers that supply built-in directory account ingest mappings.
type DirectoryAccountMappingProvider interface {
	Provider
	// DefaultDirectoryAccountMappings returns built-in mappings keyed by variant (empty string for single-mapping providers).
	DefaultDirectoryAccountMappings() map[string]MappingSpec
}

// MappingIndex resolves provider-registered default mapping specs by schema and variant.
// It is implemented by the integration registry and injected during server startup.
type MappingIndex interface {
	// SupportsVulnerabilityIngest reports whether the provider has any registered vulnerability mappings.
	SupportsVulnerabilityIngest(provider ProviderType) bool
	// DefaultVulnerabilityMapping returns the mapping spec for a provider and alert variant.
	DefaultVulnerabilityMapping(provider ProviderType, variant string) (MappingSpec, bool)
	// SupportsDirectoryAccountIngest reports whether the provider has any registered directory account mappings.
	SupportsDirectoryAccountIngest(provider ProviderType) bool
	// DefaultDirectoryAccountMapping returns the mapping spec for a provider and variant.
	DefaultDirectoryAccountMapping(provider ProviderType, variant string) (MappingSpec, bool)
}
