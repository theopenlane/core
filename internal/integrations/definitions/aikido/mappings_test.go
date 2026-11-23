package aikido

import (
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/utils/mapx"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/mappingtest"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

func prunedExample(t *testing.T, name string) json.RawMessage {
	t.Helper()

	var example any
	assert.NilError(t, json.Unmarshal(mappingtest.LoadExample(t, "examples", name), &example))

	if records, ok := example.([]any); ok {
		example = records[0]
	}

	record, ok := example.(map[string]any)
	assert.Assert(t, ok, "example %s holds no record", name)

	payload, err := json.Marshal(mapx.PruneMapZeroAny(record))
	assert.NilError(t, err)

	return payload
}

func TestMappingExpressionsValid(t *testing.T) {
	def, err := Builder(Config{})()
	assert.NilError(t, err)

	mappingtest.AssertExpressionsValid(t, def.Mappings)
}

func TestVulnerabilityMapping(t *testing.T) {
	def, err := Builder(Config{})()
	assert.NilError(t, err)

	spec := mappingtest.MappingSpec(t, def.Mappings, entityops.SchemaVulnerability.Name)
	envelope := types.MappingEnvelope{Resource: "example-repo", Payload: prunedExample(t, "exposure/issue_export_open_source.json")}

	assert.Check(t, mappingtest.AssertFiltered(t, spec, envelope))

	mapped := mappingtest.EvalMap(t, spec, envelope)

	assert.Check(t, is.Equal("1", mapped["external_id"]))
	assert.Check(t, is.Equal("CVE-2021-23337", mapped["cve_id"]))
	assert.Check(t, is.Equal("CVE-2021-23337", mapped["display_name"]))
	assert.Check(t, is.Equal("Prototype pollution in lodash", mapped["summary"]))
	assert.Check(t, is.Equal("lodash before 4.17.21 is vulnerable to command injection via the template function.", mapped["description"]))
	assert.Check(t, is.Equal("CRITICAL", mapped["severity"]))
	assert.Check(t, is.Equal(9.4, mapped["score"]))
	assert.Check(t, is.Equal("lodash", mapped["package_name"]))
	assert.Check(t, is.Equal("package-lock.json", mapped["manifest_path"]))
	assert.Check(t, is.Equal(true, mapped["fix_available"]))
	assert.Check(t, is.Equal("4.17.21", mapped["first_patched_version"]))
	assert.Check(t, is.Equal("example-repo", mapped["external_owner_id"]))
	assert.Check(t, is.Equal("https://app.aikido.dev/issues/11/detail#CVE-2021-23337", mapped["external_uri"]))
	assert.Check(t, is.Equal("open_source", mapped["category"]))
	assert.Check(t, is.Equal(true, mapped["open"]))
	assert.Check(t, is.Equal("open", mapped["vulnerability_status_name"]))
	assert.Check(t, is.Equal("2023-11-20T14:03:25Z", mapped["discovered_at"]))
	assert.Check(t, is.Nil(mapped["dismissed_at"]))
	assert.Check(t, is.Nil(mapped["fixed_at"]))
	assert.Check(t, is.Equal(float64(5), mapped["remediation_sla"]))
	assert.Check(t, is.DeepEqual([]any{"CWE-1321"}, mapped["cwe_ids"]))
	assert.Check(t, is.Nil(mapped["references"]))
	assert.Check(t, is.Equal(mappingtest.Installation.Name, mapped["source"]))

	metadata, ok := mapped["metadata"].(map[string]any)
	assert.Check(t, ok)
	assert.Check(t, is.Equal("actively_exploited", metadata["exploitability"]))
	assert.Check(t, is.Equal("4.17.15", metadata["installedVersion"]))
	assert.Check(t, is.Equal("Upgrade lodash to 4.17.21 or later.", metadata["howToFix"]))
}

func TestVulnerabilityDescriptionFallbacks(t *testing.T) {
	def, err := Builder(Config{})()
	assert.NilError(t, err)

	spec := mappingtest.MappingSpec(t, def.Mappings, entityops.SchemaVulnerability.Name)

	tests := []struct {
		name                string
		payload             string
		expectedSummary     string
		expectedDescription any
	}{
		{name: "group description only", payload: `{"id":1,"type":"open_source","group":{"description":"Attacker can trigger DOS-attack via regex"}}`, expectedSummary: "Attacker can trigger DOS-attack via regex", expectedDescription: nil},
		{name: "cve keeps group description", payload: `{"id":2,"type":"open_source","cve_id":"CVE-2026-23864","rule":"DoS in Next.js","group":{"description":"Server functions crash on crafted requests"}}`, expectedSummary: "DoS in Next.js", expectedDescription: "Server functions crash on crafted requests"},
		{name: "aikido advisory left to intel", payload: `{"id":3,"type":"open_source","cve_id":"AIKIDO-2026-161437","rule":"DoS via regex","group":{"description":"Server functions crash on crafted requests"}}`, expectedSummary: "DoS via regex", expectedDescription: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapped := mappingtest.EvalMap(t, spec, types.MappingEnvelope{Payload: json.RawMessage(tt.payload)})

			assert.Check(t, is.Equal(tt.expectedSummary, mapped["summary"]))
			assert.Check(t, is.Equal(tt.expectedDescription, mapped["description"]))
			assert.Check(t, is.Nil(mapped["published_at"]))
		})
	}
}

func TestVulnerabilityDescriptionSkipsRepeatedRule(t *testing.T) {
	def, err := Builder(Config{})()
	assert.NilError(t, err)

	spec := mappingtest.MappingSpec(t, def.Mappings, entityops.SchemaVulnerability.Name)
	mapped := mappingtest.EvalMap(t, spec, types.MappingEnvelope{Payload: json.RawMessage(`{"id":1,"type":"open_source","rule":"Attacker can trigger memory corruption","group":{"description":"Attacker can trigger memory corruption"}}`)})

	assert.Check(t, is.Equal("Attacker can trigger memory corruption", mapped["summary"]))
	assert.Check(t, is.Nil(mapped["description"]))
	assert.Check(t, is.DeepEqual([]any{}, mapped["assets_names"]))
}

func TestVulnerabilityMappingWithBulkDetails(t *testing.T) {
	def, err := Builder(Config{})()
	assert.NilError(t, err)

	spec := mappingtest.MappingSpec(t, def.Mappings, entityops.SchemaVulnerability.Name)
	mapped := mappingtest.EvalMap(t, spec, types.MappingEnvelope{Resource: "test-service", Payload: json.RawMessage(`{
		"id": 1,
		"type": "open_source",
		"cve_id": "AIKIDO-2020-123",
		"cve_id_alias": "CVE-2020-7598",
		"ignore_reasons": [
			{"kind": "manual_ignore", "reason": "Marked as false positive"},
			{"kind": "auto_ignore", "reason": "Dependency not used in production"},
			{"kind": "rule_ignore", "reason": "Path starts with examples/ in repo-1"},
			{"kind": "api_ignore", "reason": "Ignored via API"}
		]
	}`)})

	assert.Check(t, is.Equal("CVE-2020-7598", mapped["cve_id"]))
	assert.Check(t, is.Equal("AIKIDO-2020-123", mapped["display_name"]))
	assert.Check(t, is.DeepEqual([]any{"https://intel.aikido.dev/cve/AIKIDO-2020-123"}, mapped["references"]))
	assert.Check(t, is.Equal("manual_ignore", mapped["dismissed_reason"]))
	assert.Check(t, is.Equal("Marked as false positive, Dependency not used in production, Path starts with examples/ in repo-1, Ignored via API", mapped["dismissed_comment"]))

	metadata, ok := mapped["metadata"].(map[string]any)
	assert.Check(t, ok)
	assert.Check(t, is.Equal("CVE-2020-7598", metadata["cveIdAlias"]))
	assert.Check(t, is.Len(metadata["ignoreReasons"], 4))
}

func TestFindingMapping(t *testing.T) {
	def, err := Builder(Config{})()
	assert.NilError(t, err)

	spec := mappingtest.MappingSpec(t, def.Mappings, entityops.SchemaFinding.Name)
	envelope := types.MappingEnvelope{Resource: "example-repo", Payload: prunedExample(t, "exposure/issue_export_sast.json")}

	assert.Check(t, mappingtest.AssertFiltered(t, spec, envelope))

	mapped := mappingtest.EvalMap(t, spec, envelope)

	assert.Check(t, is.Equal("2", mapped["external_id"]))
	assert.Check(t, is.Equal("SQL injection", mapped["display_name"]))
	assert.Check(t, is.Equal("User input reaches a SQL query without parameterization.", mapped["description"]))
	assert.Check(t, is.Equal("Use parameterized queries.", mapped["recommended_actions"]))
	assert.Check(t, is.Equal("sast", mapped["category"]))
	assert.Check(t, is.DeepEqual([]any{"sast"}, mapped["categories"]))
	assert.Check(t, is.Equal("HIGH", mapped["severity"]))
	assert.Check(t, is.Equal(7.2, mapped["score"]))
	assert.Check(t, is.Equal(float64(72), mapped["numeric_severity"]))
	assert.Check(t, is.Equal("example-repo", mapped["resource_name"]))
	assert.Check(t, is.DeepEqual([]any{"example-repo"}, mapped["targets"]))
	assert.Check(t, is.Equal("https://app.aikido.dev/issues/12/detail", mapped["external_uri"]))
	assert.Check(t, is.DeepEqual([]any{}, mapped["references"]))
	assert.Check(t, is.Equal(false, mapped["open"]))
	assert.Check(t, is.Equal("ignored", mapped["finding_status_name"]))
	assert.Check(t, is.Equal("2023-11-20T14:03:25Z", mapped["reported_at"]))

	metadata, ok := mapped["metadata"].(map[string]any)
	assert.Check(t, ok)
	assert.Check(t, is.Equal("internal/db/query.go", metadata["affectedFile"]))
	assert.Check(t, is.Equal(float64(10), metadata["startLine"]))
}

func TestPentestFindingMapping(t *testing.T) {
	def, err := Builder(Config{})()
	assert.NilError(t, err)

	var analysis map[string]any
	assert.NilError(t, json.Unmarshal(mappingtest.LoadExample(t, "examples", "scans/attack_analysis.json"), &analysis))

	payload, err := json.Marshal(map[string]any{"id": 7, "type": "ai_pentest", "assessment_id": "f19e449d-2bf0-8d37-ab9a-3bcbc748157d", "attack_analysis": analysis})
	assert.NilError(t, err)

	spec := mappingtest.MappingSpec(t, def.Mappings, entityops.SchemaFinding.Name)
	mapped := mappingtest.EvalMap(t, spec, types.MappingEnvelope{Payload: payload})

	assert.Check(t, is.Contains(mapped["description"].(string), "/api/faq/search dynamically builds a SQL query"))
	assert.Check(t, is.Contains(mapped["recommended_actions"].(string), "Replace string interpolation with parameterized queries"))
	assert.Check(t, is.Len(mapped["steps_to_reproduce"], 2))
	assert.Check(t, is.DeepEqual([]any{"f19e449d-2bf0-8d37-ab9a-3bcbc748157d"}, mapped["scans_external_ids"]))

	metadata, ok := mapped["metadata"].(map[string]any)
	assert.Check(t, ok)
	assert.Check(t, is.Contains(metadata["attackSummary"].(string), "concatenates the user-supplied q parameter"))
}

func TestScanMapping(t *testing.T) {
	def, err := Builder(Config{})()
	assert.NilError(t, err)

	spec := mappingtest.MappingSpec(t, def.Mappings, entityops.SchemaScan.Name)
	mapped := mappingtest.EvalMap(t, spec, types.MappingEnvelope{Payload: json.RawMessage(`{"id":"f19e449d-2bf0-8d37-ab9a-3bcbc748157d","name":"Production web app pentest","status":"completed","created_at":1760000000,"kind":"ai_pentest"}`)})

	assert.Check(t, is.Equal("f19e449d-2bf0-8d37-ab9a-3bcbc748157d", mapped["external_id"]))
	assert.Check(t, is.Equal("Production web app pentest", mapped["target"]))
	assert.Check(t, is.Equal("VULNERABILITY", mapped["scan_type"]))
	assert.Check(t, is.Equal("COMPLETED", mapped["status"]))
	assert.Check(t, is.Equal("2025-10-09T08:53:20Z", mapped["scan_date"]))
	assert.Check(t, is.DeepEqual([]any{"ai_pentest"}, mapped["tags"]))
}

func TestDataleakFindingMapping(t *testing.T) {
	def, err := Builder(Config{})()
	assert.NilError(t, err)

	spec := mappingtest.VariantMappingSpec(t, def.Mappings, entityops.SchemaFinding.Name, findingVariantDataleak)
	mapped := mappingtest.EvalMap(t, spec, types.MappingEnvelope{
		Variant:  findingVariantDataleak,
		Resource: "https://www.aikido.dev",
		Payload:  json.RawMessage(`{"id":"123e4567-e89b-12d3-a456-426614174000","email":"jane.doe@example.com","status":"OPEN","last_breach_date":1705276800}`),
	})

	assert.Check(t, is.Equal("dataleak:123e4567-e89b-12d3-a456-426614174000", mapped["external_id"]))
	assert.Check(t, is.Equal("Leaked credentials for jane.doe@example.com", mapped["display_name"]))
	assert.Check(t, is.Equal(findingVariantDataleak, mapped["category"]))
	assert.Check(t, is.Equal("https://www.aikido.dev", mapped["resource_name"]))
	assert.Check(t, is.Equal(true, mapped["open"]))
	assert.Check(t, is.Equal("OPEN", mapped["finding_status_name"]))
	assert.Check(t, is.Equal("2024-01-15T00:00:00Z", mapped["reported_at"]))
}

func TestRemediationMapping(t *testing.T) {
	def, err := Builder(Config{})()
	assert.NilError(t, err)

	spec := mappingtest.MappingSpec(t, def.Mappings, entityops.SchemaRemediation.Name)

	t.Run("open pull request", func(t *testing.T) {
		mapped := mappingtest.EvalMap(t, spec, types.MappingEnvelope{Payload: json.RawMessage(`{"id":12345,"type":"dependency","status":"pr_open","autofix_target":{"code_repo_id":1,"location":"package-lock.json","title":"lodash → 4.17.21"},"triggered_at":1783119723,"finished_at":1783119840,"pull_request":{"url":"https://github.com/org/repo/pull/123","close_reason":"unknown"},"targeted_issue_ids":["1","2"]}`)})

		assert.Check(t, is.Equal("12345", mapped["external_id"]))
		assert.Check(t, is.Equal("lodash → 4.17.21", mapped["title"]))
		assert.Check(t, is.Equal("dependency", mapped["intent"]))
		assert.Check(t, is.Equal("Aikido dependency autofix in package-lock.json", mapped["instructions"]))
		assert.Check(t, is.Equal("IN_REVIEW", mapped["status"]))
		assert.Check(t, is.Equal("pr_open", mapped["state"]))
		assert.Check(t, is.Equal("", mapped["error"]))
		assert.Check(t, is.Equal("https://github.com/org/repo/pull/123", mapped["pull_request_uri"]))
		assert.Check(t, is.Equal("2026-07-03T23:04:00Z", mapped["pr_generated_at"]))
		assert.Check(t, is.Nil(mapped["completed_at"]))
		assert.Check(t, is.DeepEqual([]any{"1", "2"}, mapped["vulnerabilities_external_ids"]))
		assert.Check(t, is.DeepEqual([]any{"1", "2"}, mapped["findings_external_ids"]))
	})

	t.Run("merged", func(t *testing.T) {
		mapped := mappingtest.EvalMap(t, spec, types.MappingEnvelope{Payload: json.RawMessage(`{"id":1,"status":"pr_merged","finished_at":1783119840,"pull_request":{"url":"https://github.com/org/repo/pull/1"}}`)})

		assert.Check(t, is.Equal("COMPLETED", mapped["status"]))
		assert.Check(t, is.Equal("2026-07-03T23:04:00Z", mapped["completed_at"]))
	})

	t.Run("preview not applied", func(t *testing.T) {
		mapped := mappingtest.EvalMap(t, spec, types.MappingEnvelope{Payload: json.RawMessage(`{"id":3,"type":"sast","status":"preview_available"}`)})

		assert.Check(t, is.Equal("OPEN", mapped["status"]))
		assert.Check(t, is.Equal("Aikido sast autofix", mapped["instructions"]))
	})

	t.Run("failed", func(t *testing.T) {
		mapped := mappingtest.EvalMap(t, spec, types.MappingEnvelope{Payload: json.RawMessage(`{"id":2,"status":"autofix_failed"}`)})

		assert.Check(t, is.Equal("WONT_DO", mapped["status"]))
		assert.Check(t, is.Equal("autofix_failed", mapped["error"]))
		assert.Check(t, is.Nil(mapped["pr_generated_at"]))
	})
}

func TestCheckResultMapping(t *testing.T) {
	def, err := Builder(Config{})()
	assert.NilError(t, err)

	spec := mappingtest.MappingSpec(t, def.Mappings, entityops.SchemaCheckResult.Name)

	t.Run("failing rule", func(t *testing.T) {
		mapped := mappingtest.EvalMap(t, spec, types.MappingEnvelope{Payload: json.RawMessage(`{"framework":"soc2","standard":"SOC 2","control_code":"cc3.3","ref_code":"CC3.3","measure":"enable_threat_detection","title":"Activity Log alerts are enabled","measure_title":"Threat detection is enabled","status":"failing","report_url":"https://app.aikido.dev/reports/soc2"}`)})

		assert.Check(t, is.Equal("aikido:soc2:cc3.3:Activity Log alerts are enabled", mapped["parent_external_id"]))
		assert.Check(t, is.Equal("https://app.aikido.dev/reports/soc2", mapped["external_uri"]))
		assert.Check(t, is.Equal("Activity Log alerts are enabled", mapped["rule_title"]))
		assert.Check(t, is.Equal("FAIL", mapped["status"]))
		assert.Check(t, is.Equal("Activity Log alerts are enabled: Threat detection is enabled", mapped["details"]))
		assert.Check(t, is.DeepEqual([]any{"SOC 2", "CC3.3"}, mapped["tags"]))
	})

	t.Run("complying rule", func(t *testing.T) {
		mapped := mappingtest.EvalMap(t, spec, types.MappingEnvelope{Payload: json.RawMessage(`{"framework":"pci","standard":"PCI DSS","ref_code":"1.2","title":"RDS instances are not publicly accessible","status":"complying"}`)})

		assert.Check(t, is.Equal("PASS", mapped["status"]))
		assert.Check(t, is.Equal("", mapped["rule_title"]))
	})

	t.Run("disabled rule", func(t *testing.T) {
		mapped := mappingtest.EvalMap(t, spec, types.MappingEnvelope{Payload: json.RawMessage(`{"framework":"dora","title":"Backups are enabled","status":"disabled"}`)})

		assert.Check(t, is.Equal("UNKNOWN", mapped["status"]))
		assert.Check(t, is.DeepEqual([]any{}, mapped["tags"]))
	})
}

func TestDirectoryMappings(t *testing.T) {
	def, err := Builder(Config{})()
	assert.NilError(t, err)

	t.Run("account", func(t *testing.T) {
		spec := mappingtest.MappingSpec(t, def.Mappings, entityops.SchemaDirectoryAccount.Name)
		mapped := mappingtest.EvalMap(t, spec, types.MappingEnvelope{Resource: "1", Payload: prunedExample(t, "directory/list_users.json")})

		assert.Check(t, is.Equal("1", mapped["external_id"]))
		assert.Check(t, is.Equal("wendysnutz@gmail.com", mapped["canonical_email"]))
		assert.Check(t, is.Equal("Wendy Snutz", mapped["display_name"]))
		assert.Check(t, is.Equal("USER", mapped["account_type"]))
		assert.Check(t, is.Equal("ACTIVE", mapped["status"]))

		metadata, ok := mapped["metadata"].(map[string]any)
		assert.Check(t, ok)
		assert.Check(t, is.Equal("admin", metadata["role"]))
		assert.Check(t, is.Equal("github", metadata["authType"]))
	})

	t.Run("inactive account", func(t *testing.T) {
		spec := mappingtest.MappingSpec(t, def.Mappings, entityops.SchemaDirectoryAccount.Name)
		mapped := mappingtest.EvalMap(t, spec, types.MappingEnvelope{Payload: json.RawMessage(`{"id":2,"email":"sam@example.com"}`)})

		assert.Check(t, is.Equal("sam@example.com", mapped["display_name"]))
		assert.Check(t, is.Equal("INACTIVE", mapped["status"]))
	})

	t.Run("group", func(t *testing.T) {
		spec := mappingtest.MappingSpec(t, def.Mappings, entityops.SchemaDirectoryGroup.Name)
		mapped := mappingtest.EvalMap(t, spec, types.MappingEnvelope{Resource: "1", Payload: json.RawMessage(`{"id":1,"name":"Frontend developers","external_source":"github","external_source_id":"387505","active":true,"member_count":3}`)})

		assert.Check(t, is.Equal("1", mapped["external_id"]))
		assert.Check(t, is.Equal("Frontend developers", mapped["display_name"]))
		assert.Check(t, is.Equal("ACTIVE", mapped["status"]))
		assert.Check(t, is.Equal(float64(3), mapped["member_count"]))
	})

	t.Run("membership", func(t *testing.T) {
		spec := mappingtest.MappingSpec(t, def.Mappings, entityops.SchemaDirectoryMembership.Name)
		mapped := mappingtest.EvalMap(t, spec, types.MappingEnvelope{Resource: "1:5", Payload: json.RawMessage(`{"team_id":"1","user_id":"5"}`)})

		assert.Check(t, is.Equal("5", mapped["directory_account_id"]))
		assert.Check(t, is.Equal("1", mapped["directory_group_id"]))
		assert.Check(t, is.Equal("MEMBER", mapped["role"]))
	})
}

func TestFindingDisplayNameFallbacks(t *testing.T) {
	def, err := Builder(Config{})()
	assert.NilError(t, err)

	spec := mappingtest.MappingSpec(t, def.Mappings, entityops.SchemaFinding.Name)

	tests := []struct {
		name     string
		payload  string
		expected string
	}{
		{name: "rule", payload: `{"id":1,"type":"sast","rule":"SQL injection","group":{"title":"Injection"}}`, expected: "SQL injection"},
		{name: "group title", payload: `{"id":2,"type":"license","group":{"title":"GPL licensed dependency"}}`, expected: "GPL licensed dependency"},
		{name: "license package over group title", payload: `{"id":3,"type":"license","affected_package":"@todesktop/client-todesktop-runtime-types@1.5.7","license_type":"UNLICENSED","group":{"title":"UNLICENSED"}}`, expected: "@todesktop/client-todesktop-runtime-types@1.5.7 (UNLICENSED)"},
		{name: "type", payload: `{"id":4,"type":"license"}`, expected: "license"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapped := mappingtest.EvalMap(t, spec, types.MappingEnvelope{Payload: json.RawMessage(tt.payload)})

			assert.Check(t, is.Equal(tt.expected, mapped["display_name"]))
		})
	}
}

func TestFindingMappingWithoutResource(t *testing.T) {
	def, err := Builder(Config{})()
	assert.NilError(t, err)

	spec := mappingtest.MappingSpec(t, def.Mappings, entityops.SchemaFinding.Name)
	mapped := mappingtest.EvalMap(t, spec, types.MappingEnvelope{Payload: json.RawMessage(`{"id":3,"type":"leaked_secret"}`)})

	assert.Check(t, is.Equal("leaked_secret", mapped["display_name"]))
	assert.Check(t, is.Equal("", mapped["description"]))
	assert.Check(t, is.DeepEqual([]any{}, mapped["targets"]))
	assert.Check(t, is.Equal("", mapped["external_uri"]))
	assert.Check(t, is.Equal(false, mapped["open"]))
}

func TestAssetMappings(t *testing.T) {
	def, err := Builder(Config{})()
	assert.NilError(t, err)

	t.Run("code repository", func(t *testing.T) {
		spec := mappingtest.VariantMappingSpec(t, def.Mappings, entityops.SchemaAsset.Name, assetVariantCodeRepository)
		mapped := mappingtest.EvalMap(t, spec, types.MappingEnvelope{Variant: assetVariantCodeRepository, Resource: "Compression service", Payload: prunedExample(t, "assets/code_repository.json")})

		assert.Check(t, is.Equal("aikido:code_repository:1", mapped["source_identifier"]))
		assert.Check(t, is.Equal("Compression service", mapped["name"]))
		assert.Check(t, is.Equal("REPOSITORY", mapped["asset_type"]))
		assert.Check(t, is.Equal("R_kgDOI5RlKA", mapped["identifier"]))
		assert.Check(t, is.Equal("https://api.github.com/repos/aikidemo/compression-service", mapped["website"]))
		assert.Check(t, is.DeepEqual([]any{"production", "core"}, mapped["tags"]))
		assert.Check(t, is.Equal("IMPORTED", mapped["source_type"]))
		assert.Check(t, is.Equal("2024-07-04T08:52:43Z", mapped["observed_at"]))
		assert.Check(t, is.DeepEqual([]any{"repository", "github"}, mapped["categories"]))
		assert.Check(t, is.Equal("connected", mapped["access_model_name"]))
		assert.Check(t, is.Equal("normal", mapped["asset_data_classification_name"]))
	})

	t.Run("cloud", func(t *testing.T) {
		spec := mappingtest.VariantMappingSpec(t, def.Mappings, entityops.SchemaAsset.Name, assetVariantCloud)
		mapped := mappingtest.EvalMap(t, spec, types.MappingEnvelope{Variant: assetVariantCloud, Resource: "AWS Cloud environment", Payload: prunedExample(t, "assets/cloud.json")})

		assert.Check(t, is.Equal("aikido:cloud:1", mapped["source_identifier"]))
		assert.Check(t, is.Equal("TECHNOLOGY", mapped["asset_type"]))
		assert.Check(t, is.Equal("000000000000", mapped["identifier"]))
		assert.Check(t, is.Equal("2024-07-04T08:52:43Z", mapped["observed_at"]))
		assert.Check(t, is.DeepEqual([]any{"cloud", "aws", "production"}, mapped["categories"]))
	})

	t.Run("never scanned cloud", func(t *testing.T) {
		spec := mappingtest.VariantMappingSpec(t, def.Mappings, entityops.SchemaAsset.Name, assetVariantCloud)
		mapped := mappingtest.EvalMap(t, spec, types.MappingEnvelope{Variant: assetVariantCloud, Resource: "AWS Cloud environment", Payload: json.RawMessage(`{"id":7,"last_scanned_at":-1}`)})

		assert.Check(t, is.Nil(mapped["observed_at"]))
	})

	t.Run("domain", func(t *testing.T) {
		spec := mappingtest.VariantMappingSpec(t, def.Mappings, entityops.SchemaAsset.Name, assetVariantDomain)
		mapped := mappingtest.EvalMap(t, spec, types.MappingEnvelope{Variant: assetVariantDomain, Resource: "app.example.com", Payload: json.RawMessage(`{"id":9,"domain":"app.example.com","kind":"front_end","last_scanned_at":"1720083163"}`)})

		assert.Check(t, is.Equal("DOMAIN", mapped["asset_type"]))
		assert.Check(t, is.Equal("https://app.example.com", mapped["website"]))
		assert.Check(t, is.Equal("2024-07-04T08:52:43Z", mapped["observed_at"]))
		assert.Check(t, is.DeepEqual([]any{"domain", "front_end"}, mapped["categories"]))
		assert.Check(t, is.DeepEqual([]any{}, mapped["connected_assets_source_identifiers"]))
	})

	t.Run("domain with scheme and linked repository", func(t *testing.T) {
		spec := mappingtest.VariantMappingSpec(t, def.Mappings, entityops.SchemaAsset.Name, assetVariantDomain)
		mapped := mappingtest.EvalMap(t, spec, types.MappingEnvelope{Variant: assetVariantDomain, Resource: "https://www.aikido.dev", Payload: prunedExample(t, "assets/domains.json")})

		assert.Check(t, is.Equal("https://www.aikido.dev", mapped["website"]))
		assert.Check(t, is.DeepEqual([]any{"aikido:code_repository:1"}, mapped["connected_assets_source_identifiers"]))
	})

	t.Run("container", func(t *testing.T) {
		spec := mappingtest.VariantMappingSpec(t, def.Mappings, entityops.SchemaAsset.Name, assetVariantContainer)
		mapped := mappingtest.EvalMap(t, spec, types.MappingEnvelope{Variant: assetVariantContainer, Resource: "pied-piper/compression", Payload: prunedExample(t, "assets/containers.json")})

		assert.Check(t, is.Equal("aikido:container:1", mapped["source_identifier"]))
		assert.Check(t, is.Equal("TECHNOLOGY", mapped["asset_type"]))
		assert.Check(t, is.DeepEqual([]any{"production", "core"}, mapped["tags"]))
		assert.Check(t, is.DeepEqual([]any{"container", "aws_ecr", "debian"}, mapped["categories"]))
		assert.Check(t, is.Equal("connected", mapped["access_model_name"]))
		assert.Check(t, is.Equal("sensitive", mapped["asset_data_classification_name"]))
		assert.Check(t, is.DeepEqual([]any{"aikido:code_repository:67"}, mapped["connected_assets_source_identifiers"]))
	})

	t.Run("subdomain", func(t *testing.T) {
		spec := mappingtest.VariantMappingSpec(t, def.Mappings, entityops.SchemaAsset.Name, assetVariantSubdomain)
		mapped := mappingtest.EvalMap(t, spec, types.MappingEnvelope{Variant: assetVariantSubdomain, Resource: "app.aikido.dev", Payload: json.RawMessage(`{"id":"123e4567-e89b-12d3-a456-426614174000","name":"app.aikido.dev","dns_last_resolved_at":1772704787,"domain_id":4}`)})

		assert.Check(t, is.Equal("aikido:subdomain:123e4567-e89b-12d3-a456-426614174000", mapped["source_identifier"]))
		assert.Check(t, is.Equal("DOMAIN", mapped["asset_type"]))
		assert.Check(t, is.Equal("https://app.aikido.dev", mapped["website"]))
		assert.Check(t, is.Equal("2026-03-05T09:59:47Z", mapped["observed_at"]))
		assert.Check(t, is.DeepEqual([]any{"aikido:domain:4"}, mapped["connected_assets_source_identifiers"]))
	})

	t.Run("virtual machine", func(t *testing.T) {
		spec := mappingtest.VariantMappingSpec(t, def.Mappings, entityops.SchemaAsset.Name, assetVariantVirtualMachine)
		mapped := mappingtest.EvalMap(t, spec, types.MappingEnvelope{Variant: assetVariantVirtualMachine, Resource: "vm1", Payload: prunedExample(t, "assets/vms.json")})

		assert.Check(t, is.Equal("aikido:virtual_machine:1", mapped["source_identifier"]))
		assert.Check(t, is.Equal("nevergonnagive", mapped["identifier"]))
		assert.Check(t, is.Equal("eu-west-1", mapped["region"]))
		assert.Check(t, is.DeepEqual([]any{"virtual_machine", "aws", "production"}, mapped["categories"]))
		assert.Check(t, is.DeepEqual([]any{"aikido:cloud:1"}, mapped["connected_assets_source_identifiers"]))
	})

	t.Run("device", func(t *testing.T) {
		spec := mappingtest.VariantMappingSpec(t, def.Mappings, entityops.SchemaAsset.Name, assetVariantDevice)
		mapped := mappingtest.EvalMap(t, spec, types.MappingEnvelope{Variant: assetVariantDevice, Resource: "Yuri's MacBook Pro", Payload: prunedExample(t, "assets/devices.json")})

		assert.Check(t, is.Equal("DEVICE", mapped["asset_type"]))
		assert.Check(t, is.Equal("C02XL0GHJGH5", mapped["identifier"]))
		assert.Check(t, is.Equal("Yuri Dart", mapped["internal_owner"]))
		assert.Check(t, is.Equal("2024-05-28T00:00:00Z", mapped["observed_at"]))
		assert.Check(t, is.DeepEqual([]any{"device", "macos", "protected"}, mapped["categories"]))
	})
}
