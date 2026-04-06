package awssecurityhub

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

// TestAWSSecurityHubMappingsEvalMap verifies Security Hub findings map into vulnerability fields
func TestAWSSecurityHubMappingsEvalMap(t *testing.T) {
	envelope := types.MappingEnvelope{
		Resource: "arn:aws:ec2:us-east-1:123456789012:instance/i-1234567890",
		Payload: json.RawMessage(`{
			"Id":"arn:aws:securityhub:us-east-1::product/aws/securityhub/arn:aws:securityhub:us-east-1:123456789012:subscription/test/finding-1",
			"AwsAccountId":"123456789012",
			"CreatedAt":"2026-03-15T12:00:00Z",
			"UpdatedAt":"2026-03-15T13:00:00Z",
			"Description":"Security Hub detected a vulnerable package",
			"Title":"Critical package vulnerability",
			"Types":["Software and Configuration Checks/Vulnerabilities/CVE"],
			"Workflow":{"Status":"NEW"},
			"RecordState":"ACTIVE",
			"Severity":{"Label":"CRITICAL"},
			"Resources":[{"Id":"arn:aws:ec2:us-east-1:123456789012:instance/i-1234567890"}],
			"Vulnerabilities":[{"Id":"CVE-2026-1234","ReferenceUrls":["https://example.com/CVE-2026-1234"]}]
		}`),
	}

	raw, err := providerkit.EvalMap(context.Background(), awsSecurityHubMappings()[0].Spec.MapExpr, envelope)
	require.NoError(t, err)

	mapped, err := jsonx.ToMap(raw)
	require.NoError(t, err)

	assert.Equal(t, "arn:aws:securityhub:us-east-1::product/aws/securityhub/arn:aws:securityhub:us-east-1:123456789012:subscription/test/finding-1", mapped["externalID"])
	assert.Equal(t, "arn:aws:ec2:us-east-1:123456789012:instance/i-1234567890", mapped["externalOwnerID"])
	assert.Equal(t, "aws_security_hub", mapped["source"])
	assert.Equal(t, "Software and Configuration Checks/Vulnerabilities/CVE", mapped["category"])
	assert.Equal(t, "NEW", mapped["vulnerabilityStatusName"])
	assert.Equal(t, "CRITICAL", mapped["severity"])
	assert.Equal(t, "Critical package vulnerability", mapped["summary"])
	assert.Equal(t, "Security Hub detected a vulnerable package", mapped["description"])
	assert.Equal(t, "Critical package vulnerability", mapped["displayName"])
	assert.Equal(t, "CVE-2026-1234", mapped["cveID"])
	assert.Equal(t, "https://example.com/CVE-2026-1234", mapped["externalURI"])
	assert.Equal(t, "2026-03-15T12:00:00Z", mapped["discoveredAt"])
	assert.Equal(t, "2026-03-15T13:00:00Z", mapped["sourceUpdatedAt"])
	assert.Equal(t, true, mapped["open"])
}
