package aikido

import (
	"encoding/json"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/mappingtest"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/pkg/gala"
)

func intelExample(t *testing.T) map[string]any {
	t.Helper()

	var details map[string]any
	assert.NilError(t, json.Unmarshal(mappingtest.LoadExample(t, "examples", "exposure/cvedetails.json"), &details))

	return details
}

func TestIntelUpdateFillsAdvisory(t *testing.T) {
	vuln := &generated.Vulnerability{ID: "vuln-1", CveID: "AIKIDO-2026-10095"}

	update, changed := intelUpdate(generated.NewClient().Vulnerability.UpdateOneID(vuln.ID), vuln, intelExample(t))
	assert.Assert(t, changed)

	mutation := update.Mutation()

	description, _ := mutation.Description()
	summary, _ := mutation.Summary()
	published, _ := mutation.PublishedAt()

	assert.Check(t, is.Contains(description, "vulnerable to multiple denial of service (DoS) issues"))
	assert.Check(t, is.Equal("Attacker can trigger DOS-attack", summary))
	assert.Check(t, is.Equal(models.DateTime(time.Date(2026, 1, 27, 0, 0, 0, 0, time.UTC)), published))
}

func TestIntelUpdateKeepsSummary(t *testing.T) {
	vuln := &generated.Vulnerability{ID: "vuln-1", CveID: "AIKIDO-2026-10095", Summary: "DoS via regex"}

	update, changed := intelUpdate(generated.NewClient().Vulnerability.UpdateOneID(vuln.ID), vuln, intelExample(t))
	assert.Assert(t, changed)

	_, summarySet := update.Mutation().Summary()
	assert.Check(t, !summarySet)
}

func TestIntelUpdateWithoutDetails(t *testing.T) {
	_, changed := intelUpdate(generated.NewClient().Vulnerability.UpdateOneID("vuln-1"), &generated.Vulnerability{ID: "vuln-1"}, nil)
	assert.Check(t, !changed)
}

func TestIntelPublished(t *testing.T) {
	published, ok := intelPublished("2026-10-21T17:59:27Z")
	assert.Check(t, ok)
	assert.Check(t, is.Equal(time.Date(2026, 10, 21, 17, 59, 27, 0, time.UTC), published))

	published, ok = intelPublished("2026-01-27")
	assert.Check(t, ok)
	assert.Check(t, is.Equal(time.Date(2026, 1, 27, 0, 0, 0, 0, time.UTC), published))

	_, ok = intelPublished("")
	assert.Check(t, !ok)
}

func TestIntelEnrichDispatcherIgnoresOtherDefinitions(t *testing.T) {
	dispatcher := intelEnrichDispatcher{}

	err := dispatcher.handle(gala.HandlerContext{}, operations.IngestCompletedEnvelope{DefinitionID: "def_01K0FOSSA00000000000000001", IntegrationID: "int-1", Schemas: []string{entityops.SchemaVulnerability.Name}})
	assert.NilError(t, err)
}

func TestIntelEnrichDispatcherIgnoresOtherSchemas(t *testing.T) {
	dispatcher := intelEnrichDispatcher{}

	err := dispatcher.handle(gala.HandlerContext{}, operations.IngestCompletedEnvelope{DefinitionID: definitionID.ID(), IntegrationID: "int-1", Schemas: []string{entityops.SchemaAsset.Name}})
	assert.NilError(t, err)
}
