package gcpscc

import (
	"context"
	"os"
	"testing"

	securitycenterpb "cloud.google.com/go/securitycenter/apiv2/securitycenterpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/utils/jsonx"
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

		assert.Equal(t, "CVSS:3.1/AV:N/AC:L/PR:N/UI:R/S:C/C:H/I:L/A:N", cvssv31Vector(cvss))
	})

	t.Run("unspecified_metric", func(t *testing.T) {
		cvss := &securitycenterpb.Cvssv3{
			AttackVector:     securitycenterpb.Cvssv3_ATTACK_VECTOR_NETWORK,
			AttackComplexity: securitycenterpb.Cvssv3_ATTACK_COMPLEXITY_LOW,
		}

		assert.Equal(t, "", cvssv31Vector(cvss))
	})

	t.Run("nil", func(t *testing.T) {
		assert.Equal(t, "", cvssv31Vector(nil))
	})
}

func TestFindingConsoleURL(t *testing.T) {
	const name = "organizations/319013011373/sources/5132711838811051757/locations/global/findings/00634ea8b21d7e9db85326a5972bf191"

	const escapedName = "organizations%2F319013011373%2Fsources%2F5132711838811051757%2Flocations%2Fglobal%2Ffindings%2F00634ea8b21d7e9db85326a5972bf191"

	t.Run("project", func(t *testing.T) {
		finding := &securitycenterpb.Finding{Name: name}
		resource := &securitycenterpb.ListFindingsResponse_ListFindingsResult_Resource{
			CloudProviderMetadata: &securitycenterpb.ListFindingsResponse_ListFindingsResult_Resource_GcpMetadata{
				GcpMetadata: &securitycenterpb.GcpMetadata{ProjectDisplayName: "prod-apps-project"},
			},
		}

		assert.Equal(t, findingConsoleURLBase+escapedName+"?project=prod-apps-project", findingConsoleURL(finding, resource))
	})

	t.Run("organization_fallback", func(t *testing.T) {
		finding := &securitycenterpb.Finding{Name: name}

		assert.Equal(t, findingConsoleURLBase+escapedName+"?organizationId=319013011373", findingConsoleURL(finding, nil))
	})

	t.Run("no_name", func(t *testing.T) {
		finding := &securitycenterpb.Finding{ExternalUri: "https://example.com/finding"}

		assert.Equal(t, "https://example.com/finding", findingConsoleURL(finding, nil))
	})
}

func TestBuildFindingEnvelopeVulnerabilityExample(t *testing.T) {
	payload, err := os.ReadFile("examples/vulnerability.json")
	require.NoError(t, err)

	finding := &securitycenterpb.Finding{}
	require.NoError(t, protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(payload, finding))

	finding.ExternalUri = findingConsoleURL(finding, nil)

	envelope, err := buildFindingEnvelope("", finding, protojson.MarshalOptions{UseProtoNames: true})
	require.NoError(t, err)

	raw, err := providerkit.EvalMap(context.Background(), testMappings(t)[1].Spec.MapExpr, envelope, types.MappingInstallation{})
	require.NoError(t, err)

	mapped, err := jsonx.ToMap(raw)
	require.NoError(t, err)

	assert.Equal(t, "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:L/A:L", mapped["vector"])
	assert.Equal(t, findingConsoleURLBase+"organizations%2F521113912301%2Fsources%2F9176526532406035776%2Flocations%2Fglobal%2Ffindings%2F15989355475420362014?organizationId=521113912301", mapped["external_uri"])
}
