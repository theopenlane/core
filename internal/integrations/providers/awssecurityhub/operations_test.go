package awssecurityhub

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/integrations/helpers"
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/common/models"
)

// TestAWSSecurityHubMetadataFromPayload validates required AWS metadata parsing.
func TestAWSSecurityHubMetadataFromPayload(t *testing.T) {
	payload := types.CredentialPayload{
		Provider: types.ProviderType("aws_security_hub"),
		Data: models.CredentialSet{
			ProviderData: map[string]any{
				"region":          "us-east-1",
				"roleArn":         "arn:aws:iam::123456789012:role/SecurityHub",
				"externalId":      "external-123",
				"sessionName":     "openlane-test",
				"sessionDuration": "45m",
				"accountId":       "123456789012",
			},
		},
	}

	meta, err := awsSecurityHubMetadataFromPayload(payload)
	require.NoError(t, err)
	assert.Equal(t, "us-east-1", meta.Region)
	assert.Equal(t, "arn:aws:iam::123456789012:role/SecurityHub", meta.RoleARN)
	assert.Equal(t, "external-123", meta.ExternalID)
	assert.Equal(t, "openlane-test", meta.SessionName)
	assert.Equal(t, "123456789012", meta.AccountID)
	assert.Equal(t, 45*time.Minute, meta.SessionDuration)
}

// TestAWSSecurityHubMetadataFromPayloadMissing ensures missing metadata fails fast.
func TestAWSSecurityHubMetadataFromPayloadMissing(t *testing.T) {
	_, err := awsSecurityHubMetadataFromPayload(types.CredentialPayload{})
	assert.ErrorIs(t, err, ErrMetadataMissing)

	payload := types.CredentialPayload{Data: models.CredentialSet{ProviderData: map[string]any{}}}
	_, err = awsSecurityHubMetadataFromPayload(payload)
	assert.ErrorIs(t, err, ErrMetadataMissing)

	payload.Data.ProviderData["roleArn"] = "arn:aws:iam::123456789012:role/SecurityHub"
	_, err = awsSecurityHubMetadataFromPayload(payload)
	assert.ErrorIs(t, err, ErrRegionMissing)
}

// TestAWSSecurityHubCredentials verifies access keys are resolved from payload data.
func TestAWSSecurityHubCredentials(t *testing.T) {
	payload := types.CredentialPayload{Data: models.CredentialSet{
		AccessKeyID:     "AKIA_TEST",
		SecretAccessKey: "SECRET_TEST",
		ProviderData: map[string]any{
			"sessionToken": "session-token",
		},
	}}

	creds := helpers.AWSCredentialsFromPayload(payload)
	assert.Equal(t, "AKIA_TEST", creds.AccessKeyID)
	assert.Equal(t, "SECRET_TEST", creds.SecretAccessKey)
	assert.Equal(t, "session-token", creds.SessionToken)

	payload.Data.AccessKeyID = ""
	payload.Data.SecretAccessKey = ""
	payload.Data.ProviderData["accessKeyId"] = "AKIA_FALLBACK"
	payload.Data.ProviderData["secretAccessKey"] = "SECRET_FALLBACK"

	creds = helpers.AWSCredentialsFromPayload(payload)
	assert.Equal(t, "AKIA_FALLBACK", creds.AccessKeyID)
	assert.Equal(t, "SECRET_FALLBACK", creds.SecretAccessKey)
	assert.Equal(t, "session-token", creds.SessionToken)
}

// TestParseDuration verifies session duration parsing behavior.
func TestParseDuration(t *testing.T) {
	assert.Equal(t, time.Duration(0), helpers.ParseDuration(""))
	assert.Equal(t, time.Duration(0), helpers.ParseDuration("not-a-duration"))
	assert.Equal(t, 30*time.Minute, helpers.ParseDuration("30m"))
}
