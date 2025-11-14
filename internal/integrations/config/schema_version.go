package config

// DefaultSchemaVersion is the schema version assigned when specs omit the field
const DefaultSchemaVersion = "v1"

// supportedSchemaVersions enumerates the schema versions recognized by the loader
var supportedSchemaVersions = map[string]struct{}{
	DefaultSchemaVersion: {},
}

// supportsSchemaVersion checks if the spec declares a recognized schema version
func (s *ProviderSpec) supportsSchemaVersion() bool {
	if s == nil {
		return false
	}

	version := s.SchemaVersion
	if version == "" {
		version = DefaultSchemaVersion
	}

	_, ok := supportedSchemaVersions[version]

	return ok
}
