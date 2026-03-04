package ingest

import integrationtypes "github.com/theopenlane/core/common/integrations/types"

var schemaIngestHandlers = map[integrationtypes.MappingSchema]IngestFunc{
	integrationtypes.MappingSchemaVulnerability:    VulnerabilityIngestFunc(),
	integrationtypes.MappingSchemaDirectoryAccount: DirectoryAccountIngestFunc(),
}

// HandlerForSchema returns the ingest handler registered for a mapping schema.
func HandlerForSchema(schema integrationtypes.MappingSchema) (IngestFunc, bool) {
	handler, ok := schemaIngestHandlers[integrationtypes.NormalizeMappingSchema(schema)]

	return handler, ok
}
