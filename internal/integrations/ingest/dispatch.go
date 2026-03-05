package ingest

import (
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	integrationtypes "github.com/theopenlane/core/internal/integrations/types"
)

var schemaIngestHandlers = newSchemaIngestHandlers()

// newSchemaIngestHandlers initializes schema handler stubs from generated contracts and
// registers concrete handlers for implemented schemas.
func newSchemaIngestHandlers() map[integrationtypes.MappingSchema]IngestFunc {
	handlers := lo.SliceToMap(lo.Keys(integrationgenerated.IntegrationIngestSchemas), func(schemaName string) (integrationtypes.MappingSchema, IngestFunc) {
		return integrationtypes.MappingSchema(schemaName), nil
	})

	registerSchemaIngestHandler(handlers, integrationtypes.MappingSchemaVulnerability, VulnerabilityAlerts)
	registerSchemaIngestHandler(handlers, integrationtypes.MappingSchemaDirectoryAccount, DirectoryAccounts)

	return handlers
}

// registerSchemaIngestHandler sets the ingest handler for one schema in the supplied registry.
func registerSchemaIngestHandler(handlers map[integrationtypes.MappingSchema]IngestFunc, schema integrationtypes.MappingSchema, handler IngestFunc) {
	handlers[integrationtypes.NormalizeMappingSchema(schema)] = handler
}

// HandlerForSchema returns the ingest handler registered for a mapping schema.
func HandlerForSchema(schema integrationtypes.MappingSchema) (IngestFunc, bool) {
	handler, ok := schemaIngestHandlers[integrationtypes.NormalizeMappingSchema(schema)]

	return handler, ok && handler != nil
}
