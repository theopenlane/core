package config

const DefaultSchemaVersion = "v1"

var supportedSchemaVersions = map[string]struct{}{
	DefaultSchemaVersion: {},
}

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
