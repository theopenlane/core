package cloudflare

import (
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/internal/integrations/mappingtest"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

func TestMappingExpressionsValid(t *testing.T) {
	def, err := Builder()()
	assert.NilError(t, err)

	for _, m := range def.Mappings {
		name := m.Schema
		if m.Variant != "" {
			name += "/" + m.Variant
		}

		t.Run(name+"/filter", func(t *testing.T) {
			assert.NilError(t, providerkit.ValidateExpr(m.Spec.FilterExpr))
		})

		t.Run(name+"/map", func(t *testing.T) {
			assert.NilError(t, providerkit.ValidateExpr(m.Spec.MapExpr))
		})
	}
}

func TestSecurityCenterInsightsMapping(t *testing.T) {
	def, err := Builder()()
	assert.NilError(t, err)

	spec := mappingtest.MappingSpec(t, def.Mappings, "Finding")

	sampleInsights := mappingtest.LoadExample(t, "examples", "security_center_insights.json")

	var response struct {
		Result struct {
			Issues []json.RawMessage `json:"issues"`
		} `json:"result"`
	}

	assert.NilError(t, json.Unmarshal(sampleInsights, &response))
	assert.Assert(t, len(response.Result.Issues) > 0, "expected example to include at least one issue")

	envelope := types.MappingEnvelope{
		Resource: "6d3decf3259345241e8984cc27982f77",
		Payload:  response.Result.Issues[0],
	}

	assert.Assert(t, mappingtest.AssertFiltered(t, spec, envelope), "expected security_center_insights.json issue to pass the Finding filter")

	mapped := mappingtest.EvalMap(t, spec, envelope)

	assert.Equal(t, "8f13fada5c4fc26f2a4d795361185c7e-security_txt_not_enabled", mapped["external_id"])
	assert.Equal(t, "6d3decf3259345241e8984cc27982f77", mapped["external_owner_id"])
	assert.Equal(t, "security_txt_not_enabled", mapped["display_name"])
	assert.Equal(t, "google.com", mapped["resource_name"])
	assert.Equal(t, "configuration_suggestion", mapped["category"])
	assert.Equal(t, "Low", mapped["severity"])
	assert.Equal(t, true, mapped["open"])
	assert.Equal(t, "active", mapped["finding_status_name"])
	assert.Equal(t, "active", mapped["state"])
	assert.Equal(t, "2026-05-19T01:31:29.583623Z", mapped["event_time"])
	assert.Equal(t, "2026-04-28T01:28:27.64868Z", mapped["source_updated_at"])
	assert.Equal(t, "", mapped["recommended_actions"])
	assert.Equal(t, "", mapped["external_uri"])
	assert.DeepEqual(t, []any{"google.com"}, mapped["targets"])
	assert.DeepEqual(t, map[string]any{"affected_endpoints": []any{"google.com"}}, mapped["target_details"])
	assert.DeepEqual(t, []any{}, mapped["references"])
}

func TestDomainRegistrationsAssetMapping(t *testing.T) {
	def, err := Builder()()
	assert.NilError(t, err)

	spec := mappingtest.MappingSpec(t, def.Mappings, "Asset")
	sampleRegistrations := mappingtest.LoadExample(t, "examples", "registrar_registrations.json")

	var response struct {
		Result []json.RawMessage `json:"result"`
	}

	assert.NilError(t, json.Unmarshal(sampleRegistrations, &response))
	assert.Assert(t, len(response.Result) > 0, "expected example to include at least one registration")

	envelope := types.MappingEnvelope{
		Resource: "6d3decf3259345241e8984cc27982f77",
		Payload:  response.Result[0],
	}

	assert.Assert(t, mappingtest.AssertFiltered(t, spec, envelope), "expected registrar_registrations.json registration to pass the Asset filter")

	mapped := mappingtest.EvalMap(t, spec, envelope)

	assert.Equal(t, "theopenlane.io", mapped["source_identifier"])
	assert.Equal(t, "theopenlane.io", mapped["display_name"])
	assert.Equal(t, "theopenlane.io", mapped["name"])
	assert.Equal(t, "DOMAIN", mapped["asset_type"])
	assert.Equal(t, "2026-01-15T12:30:00Z", mapped["observed_at"])
}
