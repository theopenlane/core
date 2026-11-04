package gcpscc

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/integrations/mappingtest"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/utils/jsonx"
)

// testMappings returns the ingest mappings from the built GCP SCC definition
func testMappings(t *testing.T) []types.MappingRegistration {
	t.Helper()

	def, err := Builder("")()
	assert.NilError(t, err)

	return def.Mappings
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

		raw, err := providerkit.EvalMap(context.Background(), testMappings(t)[1].Spec.MapExpr, envelope, types.MappingInstallation{})
		assert.NilError(t, err)

		mapped, err := jsonx.ToMap(raw)
		assert.NilError(t, err)

		assert.Check(t, is.Equal("organizations/123/sources/456/findings/finding-1", mapped["external_id"]))
		assert.Check(t, is.Equal("projects/example-project/instances/vm-1", mapped["external_owner_id"]))
		assert.Check(t, is.Equal("OPEN_FIREWALL", mapped["category"]))
		assert.Check(t, is.Equal("Open", mapped["vulnerability_status_name"]))
		assert.Check(t, is.Equal("HIGH", mapped["severity"]))
		assert.Check(t, is.Equal("Firewall rule allows ingress from 0.0.0.0/0.", mapped["summary"]))
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

		raw, err := providerkit.EvalMap(context.Background(), testMappings(t)[1].Spec.MapExpr, envelope, types.MappingInstallation{})
		assert.NilError(t, err)

		mapped, err := jsonx.ToMap(raw)
		assert.NilError(t, err)

		assert.Check(t, is.Equal("", mapped["cve_id"]))
	})
}

// TestGCPSCCMappingsFindingExample tests finding mapping against examples/finding.json
func TestGCPSCCMappingsFindingExample(t *testing.T) {
	payload, err := os.ReadFile("examples/finding.json")
	assert.NilError(t, err)

	const resource = "//cloudresourcemanager.googleapis.com/projects/323616316362"

	envelope := types.MappingEnvelope{
		Resource: resource,
		Payload:  json.RawMessage(payload),
	}

	raw, err := providerkit.EvalMap(context.Background(), testMappings(t)[2].Spec.MapExpr, envelope, types.MappingInstallation{})
	assert.NilError(t, err)

	mapped, err := jsonx.ToMap(raw)
	assert.NilError(t, err)

	assert.Check(t, is.Equal("organizations/521113912301/sources/12112115738342921188/locations/global/findings/09b4bdb2ba6a4d7d910814c87e5def42", mapped["external_id"]))
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

	rawPayload, ok := mapped["raw_payload"].(map[string]any)
	assert.Assert(t, ok)
	assert.Check(t, is.Equal("organizations/521113912301/sources/12112115738342921188/locations/global/findings/09b4bdb2ba6a4d7d910814c87e5def42", rawPayload["name"]))
}

// TestGCPSCCMappingsVulnerabilityExample tests the vulnerability example payload
func TestGCPSCCMappingsVulnerabilityExample(t *testing.T) {
	payload, err := os.ReadFile("examples/vulnerability.json")
	assert.NilError(t, err)

	const resource = "//container.googleapis.com/projects/prod-project/locations/us-central1/clusters/prod-central1-main"

	envelope := types.MappingEnvelope{
		Resource: resource,
		Payload:  json.RawMessage(payload),
	}

	raw, err := providerkit.EvalMap(context.Background(), testMappings(t)[1].Spec.MapExpr, envelope, types.MappingInstallation{})
	assert.NilError(t, err)

	mapped, err := jsonx.ToMap(raw)
	assert.NilError(t, err)

	assert.Check(t, is.Equal("CVE-2025-4575", mapped["display_name"]))
	assert.Check(t, is.Equal("CVE-2025-4575", mapped["cve_id"]))
	assert.Check(t, is.Equal("organizations/521113912301/sources/9176526532406035776/locations/global/findings/15989355475420362014", mapped["external_id"]))
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
