package gcpscc

import (
	"testing"

	securitycenterpb "cloud.google.com/go/securitycenter/apiv2/securitycenterpb"
	"google.golang.org/protobuf/encoding/protojson"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated/vulnerability"
	"github.com/theopenlane/core/v2/internal/integrations/mappingtest"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
)

func TestCVSSv31Vector(t *testing.T) {
	t.Run("all_metrics", func(t *testing.T) {
		cvss := &securitycenterpb.Cvssv3{
			AttackVector:          securitycenterpb.Cvssv3_ATTACK_VECTOR_NETWORK,
			AttackComplexity:      securitycenterpb.Cvssv3_ATTACK_COMPLEXITY_LOW,
			PrivilegesRequired:    securitycenterpb.Cvssv3_PRIVILEGES_REQUIRED_NONE,
			UserInteraction:       securitycenterpb.Cvssv3_USER_INTERACTION_REQUIRED,
			Scope:                 securitycenterpb.Cvssv3_SCOPE_CHANGED,
			ConfidentialityImpact: securitycenterpb.Cvssv3_IMPACT_HIGH,
			IntegrityImpact:       securitycenterpb.Cvssv3_IMPACT_LOW,
			AvailabilityImpact:    securitycenterpb.Cvssv3_IMPACT_NONE,
		}

		assert.Check(t, is.Equal("CVSS:3.1/AV:N/AC:L/PR:N/UI:R/S:C/C:H/I:L/A:N", cvssv31Vector(cvss)))
	})

	t.Run("unspecified_metric", func(t *testing.T) {
		cvss := &securitycenterpb.Cvssv3{
			AttackVector:     securitycenterpb.Cvssv3_ATTACK_VECTOR_NETWORK,
			AttackComplexity: securitycenterpb.Cvssv3_ATTACK_COMPLEXITY_LOW,
		}

		assert.Check(t, is.Equal("", cvssv31Vector(cvss)))
	})

	t.Run("nil", func(t *testing.T) {
		assert.Check(t, is.Equal("", cvssv31Vector(nil)))
	})
}

func TestFindingConsoleURL(t *testing.T) {
	const name = "organizations/123456789012/sources/2000000000000000002/locations/global/findings/22222222222222222222222222222222"

	const escapedName = "organizations%2F123456789012%2Fsources%2F2000000000000000002%2Flocations%2Fglobal%2Ffindings%2F22222222222222222222222222222222"

	t.Run("project", func(t *testing.T) {
		finding := &securitycenterpb.Finding{Name: name}
		resource := &securitycenterpb.ListFindingsResponse_ListFindingsResult_Resource{
			CloudProviderMetadata: &securitycenterpb.ListFindingsResponse_ListFindingsResult_Resource_GcpMetadata{
				GcpMetadata: &securitycenterpb.GcpMetadata{ProjectDisplayName: "example-project"},
			},
		}

		assert.Check(t, is.Equal(findingConsoleURLBase+escapedName+"?project=example-project", findingConsoleURL(finding, resource)))
	})

	t.Run("organization_fallback", func(t *testing.T) {
		finding := &securitycenterpb.Finding{Name: name}

		assert.Check(t, is.Equal(findingConsoleURLBase+escapedName+"?organizationId=123456789012", findingConsoleURL(finding, nil)))
	})

	t.Run("no_name", func(t *testing.T) {
		finding := &securitycenterpb.Finding{ExternalUri: "https://example.com/finding"}

		assert.Check(t, is.Equal("https://example.com/finding", findingConsoleURL(finding, nil)))
	})
}

func TestBuildFindingEnvelopeVulnerabilityExample(t *testing.T) {
	payload := mappingtest.LoadExample(t, "examples", "vulnerability.json")

	finding := &securitycenterpb.Finding{}
	assert.NilError(t, protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(payload, finding))

	finding.ExternalUri = findingConsoleURL(finding, nil)

	envelope, err := buildFindingEnvelope("", finding, nil, protojson.MarshalOptions{UseProtoNames: true})
	assert.NilError(t, err)

	mapped := evalSchema(t, entityops.SchemaVulnerability, envelope)

	assert.Check(t, is.Equal("CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:L/A:L", mapped["vector"]))
	assert.Check(t, is.Equal("openssl 3.5.0-r0, fixed in 3.5.1-r0", mapped["summary"]))
	assert.Check(t, is.DeepEqual([]any{"//container.googleapis.com/projects/example-project/locations/us-central1/clusters/example-cluster"}, mapped[providerkit.EdgeExternalIDKey(entityops.SchemaVulnerability, vulnerability.EdgeAssets)]))
	assert.Check(t, is.Equal(findingConsoleURLBase+"organizations%2F123456789012%2Fsources%2F3000000000000000003%2Flocations%2Fglobal%2Ffindings%2F6000000000000000006?organizationId=123456789012", mapped["external_uri"]))
}

func TestVulnerabilitySummary(t *testing.T) {
	testCases := []struct {
		name     string
		vuln     *securitycenterpb.Vulnerability
		expected string
	}{
		{
			name:     "nil",
			vuln:     nil,
			expected: "",
		},
		{
			name: "no package name",
			vuln: &securitycenterpb.Vulnerability{
				OffendingPackage: &securitycenterpb.Package{PackageVersion: "1.0.0"},
				FixedPackage:     &securitycenterpb.Package{PackageVersion: "1.0.1"},
			},
			expected: "",
		},
		{
			name: "package only",
			vuln: &securitycenterpb.Vulnerability{
				OffendingPackage: &securitycenterpb.Package{PackageName: "openssl"},
			},
			expected: "openssl",
		},
		{
			name: "package and version without fix",
			vuln: &securitycenterpb.Vulnerability{
				OffendingPackage: &securitycenterpb.Package{PackageName: "openssl", PackageVersion: "3.5.0-r0"},
			},
			expected: "openssl 3.5.0-r0",
		},
		{
			name: "package version and fix",
			vuln: &securitycenterpb.Vulnerability{
				OffendingPackage: &securitycenterpb.Package{PackageName: "openssl", PackageVersion: "3.5.0-r0"},
				FixedPackage:     &securitycenterpb.Package{PackageVersion: "3.5.1-r0"},
			},
			expected: "openssl 3.5.0-r0, fixed in 3.5.1-r0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Check(t, is.Equal(tc.expected, vulnerabilitySummary(tc.vuln)))
		})
	}
}

func TestBuildFindingEnvelopeResourceMetadata(t *testing.T) {
	const resourceName = "//compute.googleapis.com/projects/example-project/global/targetHttpsProxies/example-proxy"

	finding := &securitycenterpb.Finding{
		Name:         "organizations/123456789012/sources/1000000000000000001/locations/global/findings/0123456789abcdef0123456789abcdef",
		ResourceName: resourceName,
		State:        securitycenterpb.Finding_ACTIVE,
		Category:     "WEAK_SSL_POLICY",
		FindingClass: securitycenterpb.Finding_MISCONFIGURATION,
	}

	resource := &securitycenterpb.ListFindingsResponse_ListFindingsResult_Resource{
		Name:        resourceName,
		DisplayName: "example-proxy",
		Type:        "google.compute.TargetHttpsProxy",
		Location:    "global",
		CloudProviderMetadata: &securitycenterpb.ListFindingsResponse_ListFindingsResult_Resource_GcpMetadata{
			GcpMetadata: &securitycenterpb.GcpMetadata{ProjectDisplayName: "example-project"},
		},
	}

	t.Run("with resource", func(t *testing.T) {
		envelope, err := buildFindingEnvelope("", finding, resource, protojson.MarshalOptions{UseProtoNames: true})
		assert.NilError(t, err)

		mapped := evalSchema(t, entityops.SchemaFinding, envelope)

		assert.Check(t, is.Equal(resourceName, mapped["resource_name"]))
		assert.Check(t, is.Equal("Weak SSL policy", mapped["display_name"]))
		assert.Check(t, is.Equal("WEAK_SSL_POLICY", mapped["category"]))
		assert.Check(t, is.DeepEqual([]any{resourceName}, mapped["targets"]))

		targetDetails, ok := mapped["target_details"].(map[string]any)
		assert.Assert(t, ok)
		assert.Check(t, is.Equal("example-proxy", targetDetails["display_name"]))
		assert.Check(t, is.Equal("google.compute.TargetHttpsProxy", targetDetails["type"]))
	})

	t.Run("asset keys", func(t *testing.T) {
		envelope, err := buildFindingEnvelope("", finding, resource, protojson.MarshalOptions{UseProtoNames: true})
		assert.NilError(t, err)

		for _, schema := range []*entityops.Schema{entityops.SchemaFinding, entityops.SchemaVulnerability, entityops.SchemaRisk} {
			mapped := evalSchema(t, schema, envelope)

			assert.Check(t, is.DeepEqual([]any{resourceName}, mapped[providerkit.EdgeExternalIDKey(schema, vulnerability.EdgeAssets)]))
		}
	})

	t.Run("without resource", func(t *testing.T) {
		envelope, err := buildFindingEnvelope("", finding, nil, protojson.MarshalOptions{UseProtoNames: true})
		assert.NilError(t, err)

		mapped := evalSchema(t, entityops.SchemaFinding, envelope)

		assert.Check(t, is.DeepEqual([]any{resourceName}, mapped["targets"]))
		assert.Check(t, is.Nil(mapped["target_details"]))
	})
}
