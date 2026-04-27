package gcpscc

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

func TestMappingExpressionsValid(t *testing.T) {
	for _, m := range gcpsccMappings() {
		name := m.Schema
		if m.Variant != "" {
			name += "/" + m.Variant
		}

		t.Run(name+"/filter", func(t *testing.T) {
			assert.NoError(t, providerkit.ValidateExpr(m.Spec.FilterExpr))
		})

		t.Run(name+"/map", func(t *testing.T) {
			assert.NoError(t, providerkit.ValidateExpr(m.Spec.MapExpr))
		})
	}
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

		raw, err := providerkit.EvalMap(context.Background(), gcpsccMappings()[1].Spec.MapExpr, envelope)
		require.NoError(t, err)

		mapped, err := jsonx.ToMap(raw)
		require.NoError(t, err)

		assert.Equal(t, "organizations/123/sources/456/findings/finding-1", mapped["externalID"])
		assert.Equal(t, "projects/example-project/instances/vm-1", mapped["externalOwnerID"])
		assert.Equal(t, "OPEN_FIREWALL", mapped["category"])
		assert.Equal(t, "Open", mapped["vulnerabilityStatusName"])
		assert.Equal(t, "HIGH", mapped["severity"])
		assert.Equal(t, "Firewall rule allows ingress from 0.0.0.0/0.", mapped["summary"])
		assert.Equal(t, "Firewall rule allows ingress from 0.0.0.0/0.", mapped["description"])
		assert.Equal(t, "CVE-2026-0001", mapped["displayName"])
		assert.Equal(t, "CVE-2026-0001", mapped["cveID"])
		assert.Equal(t, "https://console.cloud.google.com/security/command-center/findings/finding-1", mapped["externalURI"])
		assert.Equal(t, "2026-03-14T09:00:00Z", mapped["discoveredAt"])
		assert.Equal(t, "2026-03-15T12:00:00Z", mapped["sourceUpdatedAt"])

		rawPayload, ok := mapped["rawPayload"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "organizations/123/sources/456/findings/finding-1", rawPayload["name"])
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

		raw, err := providerkit.EvalMap(context.Background(), gcpsccMappings()[1].Spec.MapExpr, envelope)
		require.NoError(t, err)

		mapped, err := jsonx.ToMap(raw)
		require.NoError(t, err)

		assert.Equal(t, "", mapped["cveID"])
	})
}

// TestGCPSCCMappingsFindingExample tests the finding mapping schema
// against the real example payload in examples/finding.json
func TestGCPSCCMappingsFindingExample(t *testing.T) {
	payload, err := os.ReadFile("examples/finding.json")
	require.NoError(t, err)

	// resource matches resource_name in the example JSON
	const resource = "//cloudresourcemanager.googleapis.com/projects/323616316362"

	envelope := types.MappingEnvelope{
		Resource: resource,
		Payload:  json.RawMessage(payload),
	}

	// mappings[2] is the finding schema
	raw, err := providerkit.EvalMap(context.Background(), gcpsccMappings()[2].Spec.MapExpr, envelope)
	require.NoError(t, err)

	mapped, err := jsonx.ToMap(raw)
	require.NoError(t, err)

	assert.Equal(t, "organizations/521113912301/sources/12112115738342921188/locations/global/findings/09b4bdb2ba6a4d7d910814c87e5def42", mapped["externalID"])
	assert.Equal(t, resource, mapped["externalOwnerID"])
	assert.Equal(t, "Persistence: New API Method", mapped["category"])
	assert.Equal(t, "THREAT", mapped["findingClass"])
	assert.Equal(t, "Open", mapped["findingStatusName"])
	assert.Equal(t, true, mapped["open"])
	assert.Equal(t, "LOW", mapped["severity"])
	assert.Equal(t, "Persistence: New API Method", mapped["displayName"])
	assert.Equal(t, "2025-06-22T03:11:22.561Z", mapped["reportedAt"])
	assert.Equal(t, "2025-06-22T03:11:21.867Z", mapped["sourceUpdatedAt"])
	assert.Equal(t, "ACTIVE", mapped["state"])

	rawPayload, ok := mapped["rawPayload"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "organizations/521113912301/sources/12112115738342921188/locations/global/findings/09b4bdb2ba6a4d7d910814c87e5def42", rawPayload["name"])
}

// TestGCPSCCMappingsVulnerabilityExample tests the vulnerability mapping schema
// against the real example payload in examples/vulnerability.json
func TestGCPSCCMappingsVulnerabilityExample(t *testing.T) {
	payload, err := os.ReadFile("examples/vulnerability.json")
	require.NoError(t, err)

	// resource matches resource_name in the example JSON
	const resource = "//container.googleapis.com/projects/prod-project/locations/us-central1/clusters/prod-central1-main"

	envelope := types.MappingEnvelope{
		Resource: resource,
		Payload:  json.RawMessage(payload),
	}

	// mappings[1] is the vulnerability schema
	raw, err := providerkit.EvalMap(context.Background(), gcpsccMappings()[1].Spec.MapExpr, envelope)
	require.NoError(t, err)

	mapped, err := jsonx.ToMap(raw)
	require.NoError(t, err)

	assert.Equal(t, "CVE-2025-4575", mapped["displayName"])
	assert.Equal(t, "CVE-2025-4575", mapped["cveID"])
	assert.Equal(t, "organizations/521113912301/sources/9176526532406035776/locations/global/findings/15989355475420362014", mapped["externalID"])
	assert.Equal(t, resource, mapped["externalOwnerID"])
	assert.Equal(t, "GKE_RUNTIME_OS_VULNERABILITY", mapped["category"])
	assert.Equal(t, "Closed", mapped["vulnerabilityStatusName"])
	assert.Equal(t, false, mapped["open"])
	assert.Equal(t, "", mapped["severity"])
	assert.Equal(t, "2025-07-03T23:55:56.581Z", mapped["discoveredAt"])
	assert.Equal(t, "2025-07-10T03:12:40.174Z", mapped["sourceUpdatedAt"])
	assert.Equal(t, float64(6.5), mapped["score"])
	assert.Equal(t, "ATTACK_VECTOR_NETWORK", mapped["vector"])
	assert.Equal(t, "RUNTIME", mapped["dependencyScope"])
	assert.Equal(t, true, mapped["fixAvailable"])
	assert.Equal(t, "3.5.1-r0", mapped["firstPatchedVersion"])
	assert.Equal(t, "3.5.0-r0", mapped["vulnerableVersionRange"])
	assert.Equal(t, "openssl", mapped["packageName"])
}
