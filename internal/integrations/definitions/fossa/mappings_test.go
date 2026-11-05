package fossa

import (
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/internal/integrations/mappingtest"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
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

// TestNullArrayPayloads guards against CEL "no such overload: size" errors that occur when
// array and object fields like projects, cwes, references or source are present in the payload
// but carry an explicit null value rather than being absent.
func TestNullArrayPayloads(t *testing.T) {
	def, err := Builder()()
	assert.NilError(t, err)

	vulnSpec := mappingtest.MappingSpec(t, def.Mappings, "Vulnerability")
	findingSpec := mappingtest.MappingSpec(t, def.Mappings, "Finding")

	t.Run("vulnerability_null_fields", func(t *testing.T) {
		payload, err := json.Marshal(map[string]any{
			"id":                    20062524,
			"type":                  "vulnerability",
			"source":                nil,
			"projects":              nil,
			"cwes":                  nil,
			"references":            nil,
			"affectedVersionRanges": nil,
			"remediation":           nil,
			"epss":                  nil,
			"statuses":              nil,
			"depths":                nil,
			"severity":              nil,
			"cvss":                  nil,
		})
		assert.NilError(t, err)

		mapped := mappingtest.EvalMap(t, vulnSpec, types.MappingEnvelope{Payload: json.RawMessage(payload)})

		assert.Equal(t, "20062524", mapped["external_id"])
		assert.Equal(t, "", mapped["cve_id"])
		assert.Equal(t, "", mapped["severity"])
		assert.Equal(t, float64(0), mapped["score"])
		assert.Equal(t, float64(0), mapped["exploitability"])
		assert.Equal(t, "", mapped["package_name"])
		assert.Equal(t, "", mapped["package_ecosystem"])
		assert.Equal(t, "", mapped["vulnerable_version_range"])
		assert.Equal(t, false, mapped["fix_available"])
		assert.Equal(t, "", mapped["first_patched_version"])
		assert.Equal(t, "TRANSITIVE", mapped["dependency_scope"])
		assert.Equal(t, false, mapped["open"])
		assert.DeepEqual(t, []any{}, mapped["cwe_ids"])
		assert.DeepEqual(t, []any{}, mapped["references"])
		assert.DeepEqual(t, map[string]any{}, mapped["metadata"].(map[string]any)["remediation"])
	})

	t.Run("finding_null_fields", func(t *testing.T) {
		payload, err := json.Marshal(map[string]any{
			"id":       20012351,
			"type":     "policy_flag",
			"source":   nil,
			"projects": nil,
			"statuses": nil,
			"license":  nil,
		})
		assert.NilError(t, err)

		mapped := mappingtest.EvalMap(t, findingSpec, types.MappingEnvelope{Payload: json.RawMessage(payload)})

		assert.Equal(t, "20012351", mapped["external_id"])
		assert.Equal(t, "policy_flag", mapped["category"])
		assert.Equal(t, "", mapped["resource_name"])
		assert.Equal(t, false, mapped["open"])
		assert.DeepEqual(t, []any{}, mapped["targets"])
		assert.DeepEqual(t, map[string]any{}, mapped["target_details"])
	})
}

// TestIntegralCvssScore guards the double() coercion on cvss. FOSSA reports whole-number scores
// as JSON integers, and the mapping layer normalizes whole floats to integers before evaluation,
// so an uncoerced expression would fail to produce a float64 for the score field.
func TestIntegralCvssScore(t *testing.T) {
	def, err := Builder()()
	assert.NilError(t, err)

	vulnSpec := mappingtest.MappingSpec(t, def.Mappings, "Vulnerability")

	payload, err := json.Marshal(map[string]any{
		"id":   1,
		"cvss": 10,
		"epss": map[string]any{"score": 0},
	})
	assert.NilError(t, err)

	mapped := mappingtest.EvalMap(t, vulnSpec, types.MappingEnvelope{Payload: json.RawMessage(payload)})

	assert.Equal(t, float64(10), mapped["score"])
	assert.Equal(t, float64(0), mapped["exploitability"])
}

func TestExamplePayloads(t *testing.T) {
	def, err := Builder()()
	assert.NilError(t, err)

	vulnSpec := mappingtest.MappingSpec(t, def.Mappings, "Vulnerability")
	findingSpec := mappingtest.MappingSpec(t, def.Mappings, "Finding")

	t.Run("vulnerability_json", func(t *testing.T) {
		envelope := types.MappingEnvelope{
			Resource: "git+github.com/example-org/example-repo",
			Payload:  mappingtest.LoadExample(t, "examples", "vulnerability.json"),
		}

		assert.Assert(t, mappingtest.AssertFiltered(t, vulnSpec, envelope), "expected vulnerability.json to pass the Vulnerability filter")

		mapped := mappingtest.EvalMap(t, vulnSpec, envelope)

		assert.Equal(t, "20062524", mapped["external_id"])
		assert.Equal(t, "CVE-2025-69873", mapped["cve_id"])
		assert.Equal(t, "CVE-2025-69873", mapped["display_name"])
		assert.Equal(t, "Inefficient Regular Expression Complexity", mapped["summary"])
		assert.Equal(t, "LOW", mapped["severity"])
		assert.Equal(t, 2.9, mapped["score"])
		assert.Equal(t, "CVSS:3.1/AV:L/AC:H/PR:N/UI:N/S:U/C:N/I:N/A:L", mapped["vector"])
		assert.Equal(t, 0.00492, mapped["exploitability"])
		assert.Equal(t, "ajv", mapped["package_name"])
		assert.Equal(t, "npm", mapped["package_ecosystem"])
		assert.Equal(t, "<6.14.0, >=7.0.0-alpha.0,<8.18.0", mapped["vulnerable_version_range"])
		assert.Equal(t, true, mapped["fix_available"])
		assert.Equal(t, "6.14.0", mapped["first_patched_version"])
		assert.Equal(t, "TRANSITIVE", mapped["dependency_scope"])
		assert.Equal(t, "vulnerability", mapped["category"])
		assert.Equal(t, true, mapped["open"])
		assert.Equal(t, "ACTIVE", mapped["vulnerability_status_name"])
		assert.Equal(t, "git+github.com/example-org/example-repo", mapped["external_owner_id"])
		assert.Equal(t, "https://app.fossa.com/issues/vulnerability/20062524", mapped["external_uri"])
		assert.Equal(t, "2026-02-11T19:15:50.000Z", mapped["published_at"])
		assert.Equal(t, "2026-08-18T04:33:32.561Z", mapped["discovered_at"])
		assert.Equal(t, "2026-08-18T04:33:33.54962+00:00", mapped["source_updated_at"])
		assert.Equal(t, "FOSSA", mapped["source"])
		assert.DeepEqual(t, []any{"CWE-1333", "CWE-400"}, mapped["cwe_ids"])

		// remediation guidance is a ticket requirement: the complete fix is promoted to a first
		// class field, and the partial fix plus upgrade distances are preserved in metadata
		metadata, ok := mapped["metadata"].(map[string]any)
		assert.Assert(t, ok, "expected metadata to be an object")

		remediation, ok := metadata["remediation"].(map[string]any)
		assert.Assert(t, ok, "expected metadata.remediation to be an object")

		assert.Equal(t, "6.14.0", remediation["completeFix"])
		assert.Equal(t, "6.14.0", remediation["partialFix"])
		assert.Equal(t, "MINOR", remediation["completeFixDistance"])
		assert.Equal(t, "MINOR", remediation["partialFixDistance"])

		assert.Equal(t, "CVE-2025-69873_npm+ajv", metadata["fossaVulnId"])
		assert.Equal(t, "COMPLETED", metadata["cveStatus"])
		assert.Equal(t, "UNKNOWN", metadata["fossaExploitability"])

		epss, ok := metadata["epss"].(map[string]any)
		assert.Assert(t, ok, "expected metadata.epss to be an object")
		assert.Equal(t, 0.40057, epss["percentile"])
	})

	t.Run("finding_json", func(t *testing.T) {
		envelope := types.MappingEnvelope{
			Resource: "git+github.com/example-org/example-repo",
			Payload:  mappingtest.LoadExample(t, "examples", "finding.json"),
		}

		assert.Assert(t, mappingtest.AssertFiltered(t, findingSpec, envelope), "expected finding.json to pass the Finding filter")

		mapped := mappingtest.EvalMap(t, findingSpec, envelope)

		assert.Equal(t, "20012351", mapped["external_id"])
		assert.Equal(t, "AGPL-3.0-only in github.com/fumiama/go-docx", mapped["display_name"])
		assert.Equal(t, "policy_flag", mapped["category"])
		assert.Equal(t, "github.com/fumiama/go-docx", mapped["resource_name"])
		assert.Equal(t, "git+github.com/example-org/example-repo", mapped["external_owner_id"])
		assert.Equal(t, "https://app.fossa.com/issues/licensing/20012351", mapped["external_uri"])
		assert.Equal(t, true, mapped["open"])
		assert.Equal(t, "ACTIVE", mapped["finding_status_name"])
		assert.Equal(t, "2026-08-16T03:49:42.532Z", mapped["reported_at"])
		assert.Equal(t, "FOSSA", mapped["source"])
		assert.DeepEqual(t, []any{"policy_flag"}, mapped["categories"])
		assert.DeepEqual(t, []any{"git+github.com/example-org/example-repo"}, mapped["targets"])
	})
}
