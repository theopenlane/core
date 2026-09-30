package githubapp

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/internal/integrations/mappingtest"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// testMappings returns the ingest mappings from the built GitHub App definition
func testMappings(t *testing.T) []types.MappingRegistration {
	t.Helper()

	def, err := Builder(Config{})()
	assert.NilError(t, err)

	return def.Mappings
}

func TestMappingExpressionsValid(t *testing.T) {
	mappingtest.AssertExpressionsValid(t, testMappings(t))
}
