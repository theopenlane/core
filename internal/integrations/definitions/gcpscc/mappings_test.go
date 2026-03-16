package gcpscc

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

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

		raw, err := providerkit.EvalMap(context.Background(), gcpsccMappings()[0].Spec.MapExpr, envelope)
		require.NoError(t, err)

		mapped, err := jsonx.ToMap(raw)
		require.NoError(t, err)

		assert.Equal(t, "organizations/123/sources/456/findings/finding-1", mapped["externalID"])
		assert.Equal(t, "projects/example-project/instances/vm-1", mapped["externalOwnerID"])
		assert.Equal(t, "OPEN_FIREWALL", mapped["category"])
		assert.Equal(t, "ACTIVE", mapped["status"])
		assert.Equal(t, "HIGH", mapped["severity"])
		assert.Equal(t, "OPEN_FIREWALL", mapped["summary"])
		assert.Equal(t, "Firewall rule allows ingress from 0.0.0.0/0.", mapped["description"])
		assert.Equal(t, "google.compute.Instance projects/example-project/instances/vm-1", mapped["displayName"])
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

		raw, err := providerkit.EvalMap(context.Background(), gcpsccMappings()[0].Spec.MapExpr, envelope)
		require.NoError(t, err)

		mapped, err := jsonx.ToMap(raw)
		require.NoError(t, err)

		assert.Equal(t, "", mapped["cveID"])
	})
}
