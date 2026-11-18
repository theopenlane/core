package gcpscc

import (
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated/risk"
	"github.com/theopenlane/core/v2/internal/integrations/mappingtest"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// testMappings returns the ingest mappings from the built GCP SCC definition
func testMappings(t *testing.T) []types.MappingRegistration {
	t.Helper()

	def, err := Builder("")()
	assert.NilError(t, err)

	return def.Mappings
}

// evalSchema evaluates the GCP SCC mapping for schema against the envelope
func evalSchema(t *testing.T, schema *entityops.Schema, envelope types.MappingEnvelope) map[string]any {
	t.Helper()

	return mappingtest.EvalMap(t, mappingtest.MappingSpec(t, testMappings(t), schema.Name), envelope)
}

func TestMappingExpressionsValid(t *testing.T) {
	mappingtest.AssertExpressionsValid(t, testMappings(t))
}

// TestGCPSCCMappingsEvalMap verifies SCC finding payloads map into vulnerability fields
func TestGCPSCCMappingsEvalMap(t *testing.T) {
	t.Run("with_cve_details", func(t *testing.T) {
		envelope := types.MappingEnvelope{
			Resource: "projects/example-project/instances/vm-1",
			Payload: json.RawMessage(`{
				"name":"organizations/123/sources/456/findings/finding-1",
				"canonical_name":"google.compute.Instance projects/example-project/instances/vm-1",
				"parent":"organizations/123/sources/456",
				"resource_name":"projects/example-project/instances/vm-1",
				"state":"ACTIVE",
				"category":"OPEN_FIREWALL",
				"external_uri":"https://console.cloud.google.com/security/command-center/findings/finding-1",
				"source_properties":{"asset_type":"compute.googleapis.com/Instance"},
				"event_time":"2026-03-15T12:00:00Z",
				"create_time":"2026-03-14T09:00:00Z",
				"severity":"HIGH",
				"mute":"UNMUTED",
				"finding_class":"VULNERABILITY",
				"description":"Firewall rule allows ingress from 0.0.0.0/0.",
				"vulnerability":{
					"cve":{
						"id":"CVE-2026-0001",
						"references":[
							{"source":"NVD","uri":"https://nvd.nist.gov/vuln/detail/CVE-2026-0001"},
							{"source":"Vendor","uri":"https://example.com/advisories/CVE-2026-0001"}
						]
					}
				}
			}`),
		}

		mapped := evalSchema(t, entityops.SchemaVulnerability, envelope)

		assert.Check(t, is.Equal("organizations/123/sources/456/findings/finding-1", mapped["external_id"]))
		assert.Check(t, is.Equal("projects/example-project/instances/vm-1", mapped["external_owner_id"]))
		assert.Check(t, is.Equal("OPEN_FIREWALL", mapped["category"]))
		assert.Check(t, is.Equal("Open", mapped["vulnerability_status_name"]))
		assert.Check(t, is.Equal("HIGH", mapped["severity"]))
		assert.Check(t, is.Equal("", mapped["summary"]))
		assert.Check(t, is.Equal("Firewall rule allows ingress from 0.0.0.0/0.", mapped["description"]))
		assert.Check(t, is.Equal("CVE-2026-0001", mapped["display_name"]))
		assert.Check(t, is.Equal("CVE-2026-0001", mapped["cve_id"]))
		assert.Check(t, is.Equal("https://console.cloud.google.com/security/command-center/findings/finding-1", mapped["external_uri"]))
		assert.Check(t, is.Equal("2026-03-14T09:00:00Z", mapped["discovered_at"]))
		assert.Check(t, is.Equal("2026-03-15T12:00:00Z", mapped["source_updated_at"]))

		rawPayload, ok := mapped["raw_payload"].(map[string]any)
		assert.Assert(t, ok)
		assert.Check(t, is.Equal("organizations/123/sources/456/findings/finding-1", rawPayload["name"]))
	})

	t.Run("without_vulnerability_block", func(t *testing.T) {
		envelope := types.MappingEnvelope{
			Resource: "projects/example-project/buckets/example-bucket",
			Payload: json.RawMessage(`{
				"name":"organizations/123/sources/456/findings/finding-2",
				"resource_name":"projects/example-project/buckets/example-bucket",
				"state":"INACTIVE",
				"category":"PUBLIC_BUCKET",
				"create_time":"2026-03-14T09:00:00Z",
				"finding_class":"MISCONFIGURATION"
			}`),
		}

		mapped := evalSchema(t, entityops.SchemaVulnerability, envelope)

		assert.Check(t, is.Equal("", mapped["cve_id"]))
	})
}

// TestGCPSCCMappingsFindingExample tests finding mapping against examples/finding.json
func TestGCPSCCMappingsFindingExample(t *testing.T) {
	const resource = "//cloudresourcemanager.googleapis.com/projects/987654321098"

	envelope := types.MappingEnvelope{
		Resource: resource,
		Payload:  mappingtest.LoadExample(t, "examples", "finding.json"),
	}

	mapped := evalSchema(t, entityops.SchemaFinding, envelope)

	assert.Check(t, is.Equal("organizations/123456789012/sources/4000000000000000004/locations/global/findings/11111111111111111111111111111111", mapped["external_id"]))
	assert.Check(t, is.Equal(resource, mapped["external_owner_id"]))
	assert.Check(t, is.Equal("Persistence: New API Method", mapped["category"]))
	assert.Check(t, is.Equal("THREAT", mapped["finding_class"]))
	assert.Check(t, is.Equal("Open", mapped["finding_status_name"]))
	assert.Check(t, is.Equal(true, mapped["open"]))
	assert.Check(t, is.Equal("LOW", mapped["severity"]))
	assert.Check(t, is.Equal("Persistence: New API Method", mapped["display_name"]))
	assert.Check(t, is.Equal("2025-06-22T03:11:22.561Z", mapped["reported_at"]))
	assert.Check(t, is.Equal("2025-06-22T03:11:21.867Z", mapped["source_updated_at"]))
	assert.Check(t, is.Equal("ACTIVE", mapped["state"]))
	assert.Check(t, is.Equal(resource, mapped["resource_name"]))
	assert.Check(t, is.DeepEqual([]any{"https://attack.mitre.org/tactics/TA0003/"}, mapped["references"]))

	rawPayload, ok := mapped["raw_payload"].(map[string]any)
	assert.Assert(t, ok)
	assert.Check(t, is.Equal("organizations/123456789012/sources/4000000000000000004/locations/global/findings/11111111111111111111111111111111", rawPayload["name"]))
}

// TestGCPSCCMappingsVulnerabilityExample tests the vulnerability example payload
func TestGCPSCCMappingsVulnerabilityExample(t *testing.T) {
	const resource = "//container.googleapis.com/projects/example-project/locations/us-central1/clusters/example-cluster"

	envelope := types.MappingEnvelope{
		Resource: resource,
		Payload:  mappingtest.LoadExample(t, "examples", "vulnerability.json"),
	}

	mapped := evalSchema(t, entityops.SchemaVulnerability, envelope)

	assert.Check(t, is.Equal("CVE-2025-4575", mapped["display_name"]))
	assert.Check(t, is.Equal("CVE-2025-4575", mapped["cve_id"]))
	assert.Check(t, is.Equal("organizations/123456789012/sources/3000000000000000003/locations/global/findings/6000000000000000006", mapped["external_id"]))
	assert.Check(t, is.Equal(resource, mapped["external_owner_id"]))
	assert.Check(t, is.Equal("GKE_RUNTIME_OS_VULNERABILITY", mapped["category"]))
	assert.Check(t, is.Equal("Closed", mapped["vulnerability_status_name"]))
	assert.Check(t, is.Equal(false, mapped["open"]))
	assert.Check(t, is.Equal("", mapped["severity"]))
	assert.Check(t, is.Equal("2025-07-03T23:55:56.581Z", mapped["discovered_at"]))
	assert.Check(t, is.Equal("2025-07-10T03:12:40.174Z", mapped["source_updated_at"]))
	assert.Check(t, is.Equal(float64(6.5), mapped["score"]))
	assert.Check(t, is.Equal("", mapped["vector"]))
	assert.Check(t, is.Equal("RUNTIME", mapped["dependency_scope"]))
	assert.Check(t, is.Equal(true, mapped["fix_available"]))
	assert.Check(t, is.Equal("3.5.1-r0", mapped["first_patched_version"]))
	assert.Check(t, is.Equal("3.5.0-r0", mapped["vulnerable_version_range"]))
	assert.Check(t, is.Equal("openssl", mapped["package_name"]))
}

func TestGCPSCCMappingsFindingSourceProperties(t *testing.T) {
	envelope := types.MappingEnvelope{
		Resource: "//compute.googleapis.com/projects/example-project/global/targetHttpsProxies/example-proxy",
		Payload: json.RawMessage(`{
			"name":"organizations/123456789012/sources/1000000000000000001/locations/global/findings/0123456789abcdef0123456789abcdef",
			"resource_name":"//compute.googleapis.com/projects/example-project/global/targetHttpsProxies/example-proxy",
			"state":"ACTIVE",
			"category":"WEAK_SSL_POLICY",
			"finding_class":"MISCONFIGURATION",
			"next_steps":"Update the SSL policy",
			"attack_exposure":{"score":4.5},
			"source_properties":{
				"Recommendation":"Use a custom SSL policy with MODERN or RESTRICTED profile",
				"Explanation":"The target proxy uses the default SSL policy"
			}
		}`),
	}

	mapped := evalSchema(t, entityops.SchemaFinding, envelope)

	assert.Check(t, is.Equal("Use a custom SSL policy with MODERN or RESTRICTED profile", mapped["recommendation"]))
	assert.Check(t, is.Equal("Update the SSL policy", mapped["recommended_actions"]))
	assert.Check(t, is.DeepEqual([]any{"The target proxy uses the default SSL policy"}, mapped["steps_to_reproduce"]))
	assert.Check(t, is.Equal(4.5, mapped["exploitability"]))
	assert.Check(t, is.DeepEqual([]any{}, mapped["references"]))
}

func TestGCPSCCMappingsMute(t *testing.T) {
	testCases := []struct {
		name               string
		payload            string
		expectedOpen       bool
		expectedStatusName string
		expectedDismissed  any
		expectedReason     any
	}{
		{
			name:               "active unmuted",
			payload:            `{"name":"organizations/123456789012/sources/1000000000000000001/findings/f1","state":"ACTIVE","mute":"UNMUTED","mute_update_time":"2026-03-14T09:00:00Z"}`,
			expectedOpen:       true,
			expectedStatusName: "Open",
		},
		{
			name:               "active statically muted",
			payload:            `{"name":"organizations/123456789012/sources/1000000000000000001/findings/f1","state":"ACTIVE","mute":"MUTED","mute_update_time":"2026-03-14T09:00:00Z","mute_initiator":"organizations/123456789012/muteConfigs/example-rule"}`,
			expectedStatusName: "Dismissed",
			expectedDismissed:  "2026-03-14T09:00:00Z",
			expectedReason:     "organizations/123456789012/muteConfigs/example-rule",
		},
		{
			name:               "active dynamically muted",
			payload:            `{"name":"organizations/123456789012/sources/1000000000000000001/findings/f1","state":"ACTIVE","mute":"UNDEFINED","mute_update_time":"2026-03-14T09:00:00Z","mute_info":{"dynamic_mute_records":[{"mute_config":"organizations/123456789012/muteConfigs/example-rule"}]}}`,
			expectedStatusName: "Dismissed",
			expectedDismissed:  "2026-03-14T09:00:00Z",
		},
		{
			name:               "inactive unmuted",
			payload:            `{"name":"organizations/123456789012/sources/1000000000000000001/findings/f1","state":"INACTIVE","mute":"UNDEFINED"}`,
			expectedStatusName: "Closed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			envelope := types.MappingEnvelope{Resource: "projects/example-project", Payload: json.RawMessage(tc.payload)}

			finding := evalSchema(t, entityops.SchemaFinding, envelope)

			assert.Check(t, is.Equal(tc.expectedOpen, finding["open"]))
			assert.Check(t, is.Equal(tc.expectedStatusName, finding["finding_status_name"]))

			vuln := evalSchema(t, entityops.SchemaVulnerability, envelope)

			assert.Check(t, is.Equal(tc.expectedOpen, vuln["open"]))
			assert.Check(t, is.Equal(tc.expectedStatusName, vuln["vulnerability_status_name"]))
			assert.Check(t, is.Equal(tc.expectedDismissed, vuln["dismissed_at"]))
			assert.Check(t, is.Equal(tc.expectedReason, vuln["dismissed_reason"]))
		})
	}
}

func TestGCPSCCMappingsRisk(t *testing.T) {
	testCases := []struct {
		name           string
		payload        string
		expectedStatus string
		expectedImpact any
	}{
		{
			name:           "active medium",
			payload:        `{"name":"organizations/123456789012/sources/2000000000000000002/findings/risk-1","resource_name":"//storage.googleapis.com/example-bucket","state":"ACTIVE","severity":"MEDIUM","category":"TOXIC_COMBINATION","finding_class":"TOXIC_COMBINATION","create_time":"2026-03-14T09:00:00Z","description":"Public bucket with sensitive data","next_steps":"Remove public access"}`,
			expectedStatus: "OPEN",
			expectedImpact: "MODERATE",
		},
		{
			name:           "inactive high",
			payload:        `{"name":"organizations/123456789012/sources/2000000000000000002/findings/risk-1","resource_name":"//storage.googleapis.com/example-bucket","state":"INACTIVE","severity":"HIGH","category":"TOXIC_COMBINATION","finding_class":"TOXIC_COMBINATION","create_time":"2026-03-14T09:00:00Z","description":"Public bucket with sensitive data","next_steps":"Remove public access"}`,
			expectedStatus: "CLOSED",
			expectedImpact: "HIGH",
		},
		{
			name:           "active muted",
			payload:        `{"name":"organizations/123456789012/sources/2000000000000000002/findings/risk-1","resource_name":"//storage.googleapis.com/example-bucket","state":"ACTIVE","mute":"MUTED","severity":"HIGH","category":"TOXIC_COMBINATION","finding_class":"TOXIC_COMBINATION","create_time":"2026-03-14T09:00:00Z","description":"Public bucket with sensitive data","next_steps":"Remove public access"}`,
			expectedStatus: "ACCEPTED",
			expectedImpact: "HIGH",
		},
		{
			name:           "unspecified severity",
			payload:        `{"name":"organizations/123456789012/sources/2000000000000000002/findings/risk-1","resource_name":"//storage.googleapis.com/example-bucket","state":"ACTIVE","severity":"SEVERITY_UNSPECIFIED","category":"TOXIC_COMBINATION","finding_class":"TOXIC_COMBINATION","create_time":"2026-03-14T09:00:00Z","description":"Public bucket with sensitive data","next_steps":"Remove public access"}`,
			expectedStatus: "OPEN",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			envelope := types.MappingEnvelope{
				Resource: "projects/example-project",
				Payload:  json.RawMessage(tc.payload),
			}

			mapped := evalSchema(t, entityops.SchemaRisk, envelope)

			assert.Check(t, is.Equal("organizations/123456789012/sources/2000000000000000002/findings/risk-1", mapped["external_id"]))
			assert.Check(t, is.Equal("TOXIC_COMBINATION", mapped["name"]))
			assert.Check(t, is.Equal(tc.expectedStatus, mapped["status"]))
			assert.Check(t, is.Equal(tc.expectedImpact, mapped["impact"]))
			assert.Check(t, is.Equal("2026-03-14T09:00:00Z", mapped["observed_at"]))
			assert.Check(t, is.Equal("Public bucket with sensitive data", mapped["details"]))
			assert.Check(t, is.Equal("TOXIC_COMBINATION", mapped["risk_category_name"]))
			assert.Check(t, is.Equal("Remove public access", mapped["mitigation"]))
			assert.Check(t, is.DeepEqual([]any{"//storage.googleapis.com/example-bucket"}, mapped[providerkit.EdgeExternalIDKey(entityops.SchemaRisk, risk.EdgeAssets)]))
		})
	}
}

func TestGCPSCCMappingsRiskObservedAtMissing(t *testing.T) {
	envelope := types.MappingEnvelope{
		Resource: "projects/example-project",
		Payload:  json.RawMessage(`{"name":"organizations/123456789012/sources/2000000000000000002/findings/risk-1","state":"ACTIVE"}`),
	}

	mapped := evalSchema(t, entityops.SchemaRisk, envelope)

	assert.Check(t, is.Nil(mapped["observed_at"]))
}
