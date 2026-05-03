package awssecurityhub

import (
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/internal/integrations/mappingtest"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

func TestMappingExpressionsValid(t *testing.T) {
	all := append(awsSecurityHubMappings(), awsIamMappings()...)

	for _, m := range all {
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

// TestNullArrayPayloads guards against CEL "no such overload: size" errors that occur
// when array fields like Resources, Types, or Vulnerabilities are present in the payload
// but carry an explicit null value rather than being absent.
func TestNullArrayPayloads(t *testing.T) {
	mappings := awsSecurityHubMappings()
	findingSpec := mappingtest.MappingSpec(t, mappings, "Finding")
	vulnSpec := mappingtest.MappingSpec(t, mappings, "Vulnerability")

	t.Run("finding_null_arrays", func(t *testing.T) {
		payload, err := json.Marshal(map[string]any{
			"Id":           "test-finding-id",
			"AwsAccountId": "123456789012",
			"Types":        nil,
			"Resources":    nil,
		})
		assert.NilError(t, err)

		envelope := types.MappingEnvelope{Payload: json.RawMessage(payload)}
		mapped := mappingtest.EvalMap(t, findingSpec, envelope)

		assert.Equal(t, "", mapped["category"])
		assert.Equal(t, "", mapped["resourceName"])
		assert.DeepEqual(t, []any{}, mapped["targets"])
		assert.DeepEqual(t, map[string]any{}, mapped["targetDetails"])
		assert.Equal(t, "123456789012", mapped["externalOwnerID"])
	})

	t.Run("vulnerability_null_arrays", func(t *testing.T) {
		payload, err := json.Marshal(map[string]any{
			"Id":              "test-vuln-id",
			"AwsAccountId":    "123456789012",
			"Types":           nil,
			"Resources":       nil,
			"Vulnerabilities": nil,
		})
		assert.NilError(t, err)

		envelope := types.MappingEnvelope{Payload: json.RawMessage(payload)}
		mapped := mappingtest.EvalMap(t, vulnSpec, envelope)

		assert.Equal(t, "", mapped["category"])
		assert.Equal(t, "123456789012", mapped["externalOwnerID"])
		assert.Equal(t, "", mapped["cveID"])
		assert.Equal(t, false, mapped["fixAvailable"])
		assert.Equal(t, "", mapped["firstPatchedVersion"])
		assert.DeepEqual(t, []any{}, mapped["references"])
		assert.Equal(t, float64(0), mapped["score"])
	})
}

func TestExamplePayloads(t *testing.T) {
	mappings := awsSecurityHubMappings()
	findingSpec := mappingtest.MappingSpec(t, mappings, "Finding")
	vulnSpec := mappingtest.MappingSpec(t, mappings, "Vulnerability")

	t.Run("vulnerability_json", func(t *testing.T) {
		envelope := types.MappingEnvelope{Payload: mappingtest.LoadExample(t, "examples", "vulnerability.json")}

		assert.Assert(t, mappingtest.AssertFiltered(t, vulnSpec, envelope), "expected vulnerability.json to pass the Vulnerability filter")

		mapped := mappingtest.EvalMap(t, vulnSpec, envelope)

		assert.Equal(t, "arn:aws:inspector2:us-east-1:123456789012:finding/FINDING_ID", mapped["externalID"])
		assert.Equal(t, "CVE-2022-34918", mapped["displayName"])
		assert.Equal(t, "CVE-2022-34918", mapped["cveID"])
		assert.Equal(t, "arn:aws:ec2:us-east-1:123456789012:i-0f1ed287081bdf0fb", mapped["externalOwnerID"])
		assert.Equal(t, "Software and Configuration Checks/Vulnerabilities/CVE", mapped["category"])
		assert.Equal(t, "HIGH", mapped["severity"])
		assert.Equal(t, "CVE-2022-34918 - kernel", mapped["summary"])
		assert.Equal(t, "An issue was discovered in the Linux kernel through 5.18.9. A type confusion bug in nft_set_elem_init (leading to a buffer overflow) could be used by a local attacker to escalate privileges...", mapped["description"])
		assert.Equal(t, true, mapped["open"])
		assert.Equal(t, "NEW", mapped["vulnerabilityStatusName"])
		assert.Equal(t, float64(7.8), mapped["score"])
		assert.Equal(t, true, mapped["fixAvailable"])
		assert.Equal(t, "0:5.10.130-118.517.amzn2", mapped["firstPatchedVersion"])
		assert.Equal(t, "2023-01-31T20:25:38Z", mapped["discoveredAt"])
		assert.Equal(t, "2023-05-04T18:18:43Z", mapped["sourceUpdatedAt"])
		assert.DeepEqual(t, []any{
			"https://git.kernel.org/pub/scm/linux/kernel/git/netdev/net.git/commit/?id=7e6bc1f6cabcd30aba0b11219d8e01b952eacbb6",
			"https://lore.kernel.org/netfilter-devel/cd9428b6-7ffb-dd22-d949-d86f4869f452@randorisec.fr/T/",
			"https://www.debian.org/security/2022/dsa-5191",
		}, mapped["references"])
	})

	t.Run("finding_json", func(t *testing.T) {
		envelope := types.MappingEnvelope{Payload: mappingtest.LoadExample(t, "examples", "finding.json")}

		assert.Assert(t, mappingtest.AssertFiltered(t, findingSpec, envelope), "expected finding.json to pass the Finding filter")

		mapped := mappingtest.EvalMap(t, findingSpec, envelope)

		assert.Equal(t, "arn:aws:securityhub:us-east-1:123456789123:security-control/S3.8/finding/5441c4a1-afb5-4000-b037-c98eebdd8e40", mapped["externalID"])
		assert.Equal(t, "S3 general purpose buckets should block public access", mapped["displayName"])
		assert.Equal(t, "arn:aws:s3:::aws-cloudtrail-logs-123456789123-f4ef37f5", mapped["externalOwnerID"])
		assert.Equal(t, "Software and Configuration Checks/Industry and Regulatory Standards", mapped["category"])
		assert.Equal(t, "This control checks whether an Amazon S3 general purpose bucket blocks public access at the bucket level. The control fails if any of the following settings are set to false: ignorePublicAcls, blockPublicPolicy, blockPublicAcls, restrictPublicBuckets.", mapped["description"])
		assert.Equal(t, "INFORMATIONAL", mapped["severity"])
		assert.Equal(t, false, mapped["open"])
		assert.Equal(t, "RESOLVED", mapped["findingStatusName"])
		assert.Equal(t, "ACTIVE", mapped["state"])
		assert.Assert(t, mapped["externalURI"] == nil, "expected externalURI to be nil since SourceUrl is null in payload")
		assert.Equal(t, "2026-04-21T18:18:53.823Z", mapped["reportedAt"])
		assert.Equal(t, "2026-04-21T18:18:53.823Z", mapped["eventTime"])
		assert.Equal(t, "2026-04-24T17:40:47.398Z", mapped["sourceUpdatedAt"])
		assert.Equal(t, "https://docs.aws.amazon.com/console/securityhub/S3.8/remediation", mapped["references"].([]any)[0])
	})
}
