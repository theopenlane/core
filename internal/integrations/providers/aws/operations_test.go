package aws

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/models"
	awskit "github.com/theopenlane/core/internal/integrations/providers/awskit"
)

// TestAWSMetadataFromPayload validates required AWS metadata parsing.
func TestAWSMetadataFromPayload(t *testing.T) {
	payload := models.CredentialSet{
		ProviderData: json.RawMessage(`{
				"region":"us-east-1",
				"roleArn":"arn:aws:iam::123456789012:role/SecurityHub",
				"externalId":"external-123",
				"sessionName":"openlane-test",
				"sessionDuration":"45m",
				"accountId":"123456789012"
			}`),
	}

	meta, err := awsMetadataFromPayload(payload, awsDefaultSession)
	require.NoError(t, err)
	assert.Equal(t, "us-east-1", meta.Region)
	assert.Equal(t, "arn:aws:iam::123456789012:role/SecurityHub", meta.RoleARN)
	assert.Equal(t, "external-123", meta.ExternalID)
	assert.Equal(t, "openlane-test", meta.SessionName)
	assert.Equal(t, "123456789012", meta.AccountID)
	assert.Equal(t, 45*time.Minute, meta.SessionDuration)
}

// TestAWSMetadataFromPayloadMissing ensures missing metadata fails fast.
func TestAWSMetadataFromPayloadMissing(t *testing.T) {
	_, err := awsMetadataFromPayload(models.CredentialSet{}, awsDefaultSession)
	assert.ErrorIs(t, err, ErrMetadataMissing)

	payload := models.CredentialSet{ProviderData: json.RawMessage(`{}`)}
	_, err = awsMetadataFromPayload(payload, awsDefaultSession)
	assert.ErrorIs(t, err, ErrRoleARNMissing)

	payload.ProviderData = json.RawMessage(`{"roleArn":"arn:aws:iam::123456789012:role/SecurityHub"}`)
	_, err = awsMetadataFromPayload(payload, awsDefaultSession)
	assert.ErrorIs(t, err, ErrRegionMissing)
}

// TestAWSCredentialsFromPayload verifies access keys are resolved from payload data.
func TestAWSCredentialsFromPayload(t *testing.T) {
	payload := models.CredentialSet{
		AccessKeyID:     "AKIA_TEST",
		SecretAccessKey: "SECRET_TEST",
		SessionToken:    "session-token",
		ProviderData:    json.RawMessage(`{"sessionToken":"ignored-token"}`),
	}

	creds := awskit.AWSCredentialsFromPayload(payload)
	assert.Equal(t, "AKIA_TEST", creds.AccessKeyID)
	assert.Equal(t, "SECRET_TEST", creds.SecretAccessKey)
	assert.Equal(t, "session-token", creds.SessionToken)

	payload.AccessKeyID = ""
	payload.SecretAccessKey = ""
	payload.SessionToken = ""

	creds = awskit.AWSCredentialsFromPayload(payload)
	assert.Equal(t, "", creds.AccessKeyID)
	assert.Equal(t, "", creds.SecretAccessKey)
	assert.Equal(t, "", creds.SessionToken)
}

// TestParseDuration verifies session duration parsing behavior.
func TestParseDuration(t *testing.T) {
	assert.Equal(t, time.Duration(0), awskit.ParseDuration(""))
	assert.Equal(t, time.Duration(0), awskit.ParseDuration("not-a-duration"))
	assert.Equal(t, 30*time.Minute, awskit.ParseDuration("30m"))
}
