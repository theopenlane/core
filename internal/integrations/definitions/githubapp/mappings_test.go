package githubapp

import (
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
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

func TestCodeScanning_MappingFindings(t *testing.T) {

	mapping, found := lo.Find(testMappings(t), func(candidate types.MappingRegistration) bool {
		return candidate.Variant == githubAlertTypeCodeScanning
	})
	assert.Assert(t, found)

	assert.Equal(t, entityops.SchemaFinding.Name, mapping.Schema)

	payload := mappingtest.LoadExample(t, "examples", "code_scanning_alert.json")

	envelope := types.MappingEnvelope{
		Variant:  githubAlertTypeCodeScanning,
		Resource: "theopenlane/iam",
		Payload:  payload,
	}
	assert.Assert(t, mappingtest.AssertFiltered(t, mapping.Spec, envelope))

	mapped := mappingtest.EvalMap(t, mapping.Spec, envelope)

	assert.Equal(t, "dismissed", mapped["finding_status_name"])
	assert.Equal(t, false, mapped["open"])
	assert.Equal(t, "github:theopenlane/iam:code_scanning:35", mapped["external_id"])
	assert.Equal(t, "code_scanning", mapped["category"])
	assert.Equal(t, "high", mapped["severity"])
}
