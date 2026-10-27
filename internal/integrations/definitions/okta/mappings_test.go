package okta

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/internal/integrations/mappingtest"
)

func TestMappingExpressionsValid(t *testing.T) {
	def, err := Builder(false)()
	assert.NilError(t, err)

	mappingtest.AssertExpressionsValid(t, def.Mappings)
}
