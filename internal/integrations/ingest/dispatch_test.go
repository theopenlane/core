package ingest

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	integrationtypes "github.com/theopenlane/core/internal/integrations/types"
)

// TestImplementedSchemasHaveHandlers asserts that schemas with known ingest implementations
// have a concrete handler registered; this guards against accidental handler removal.
func TestImplementedSchemasHaveHandlers(t *testing.T) {
	implemented := []integrationtypes.MappingSchema{
		integrationtypes.MappingSchemaVulnerability,
		integrationtypes.MappingSchemaDirectoryAccount,
	}

	for _, schema := range implemented {
		_, ok := HandlerForSchema(schema)
		assert.True(t, ok, "ingest handler missing for schema %q", schema)
	}
}

// TestGeneratedSchemasHandlerCoverage logs any generated schemas that do not yet have a
// registered ingest handler. This is informational — schemas may be annotated in advance of
// their handler implementation.
func TestGeneratedSchemasHandlerCoverage(t *testing.T) {
	for schemaName := range integrationgenerated.IntegrationIngestSchemas {
		schema := integrationtypes.MappingSchema(schemaName)
		_, ok := HandlerForSchema(schema)
		if !ok {
			t.Logf("schema %q has no ingest handler registered (pending implementation)", schemaName)
		}
	}
}
