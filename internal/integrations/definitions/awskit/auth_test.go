package awskit

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetadataFromProviderData_ValidData(t *testing.T) {
	raw, err := json.Marshal(map[string]any{
		"roleArn":         "arn:aws:iam::123456789012:role/MyRole",
		"region":          "us-east-1",
		"homeRegion":      "us-east-1",
		"accountId":       "123456789012",
		"accountScope":    "all",
		"accessKeyId":     "AKIAIOSFODNN7EXAMPLE",
		"secretAccessKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		"sessionToken":    "AQoDYXdzEJr",
	})
	require.NoError(t, err)

	meta, err := MetadataFromProviderData(raw, "test-session")
	require.NoError(t, err)

	assert.Equal(t, "arn:aws:iam::123456789012:role/MyRole", meta.RoleARN)
	assert.Equal(t, "us-east-1", meta.Region)
	assert.Equal(t, "us-east-1", meta.HomeRegion)
	assert.Equal(t, "123456789012", meta.AccountID)
	assert.Equal(t, "all", meta.AccountScope)
	assert.Equal(t, "AKIAIOSFODNN7EXAMPLE", meta.AccessKeyID)
	assert.Equal(t, "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", meta.SecretAccessKey)
	assert.Equal(t, "AQoDYXdzEJr", meta.SessionToken)
}

func TestMetadataFromProviderData_DefaultSessionName(t *testing.T) {
	raw, err := json.Marshal(map[string]any{
		"roleArn": "arn:aws:iam::123:role/R",
		"region":  "us-west-2",
	})
	require.NoError(t, err)

	meta, err := MetadataFromProviderData(raw, "default-session")
	require.NoError(t, err)

	assert.Equal(t, "default-session", meta.SessionName)
}

func TestMetadataFromProviderData_SessionNameFromData(t *testing.T) {
	raw, err := json.Marshal(map[string]any{
		"roleArn":     "arn:aws:iam::123:role/R",
		"region":      "us-west-2",
		"sessionName": "custom-session",
	})
	require.NoError(t, err)

	meta, err := MetadataFromProviderData(raw, "default-session")
	require.NoError(t, err)

	assert.Equal(t, "custom-session", meta.SessionName)
}

func TestMetadataFromProviderData_RegionFallback(t *testing.T) {
	raw, err := json.Marshal(map[string]any{
		"homeRegion": "eu-central-1",
	})
	require.NoError(t, err)

	meta, err := MetadataFromProviderData(raw, "session")
	require.NoError(t, err)

	assert.Equal(t, "eu-central-1", meta.Region)
	assert.Equal(t, "eu-central-1", meta.HomeRegion)
}

func TestMetadataFromProviderData_DefaultAccountScope(t *testing.T) {
	raw, err := json.Marshal(map[string]any{
		"region": "us-east-1",
	})
	require.NoError(t, err)

	meta, err := MetadataFromProviderData(raw, "session")
	require.NoError(t, err)

	assert.Equal(t, AccountScopeAll, meta.AccountScope)
}

func TestMetadataFromProviderData_EmptyInput(t *testing.T) {
	meta, err := MetadataFromProviderData(nil, "session")
	require.NoError(t, err)

	assert.Equal(t, AccountScopeAll, meta.AccountScope)
	assert.Equal(t, "session", meta.SessionName)
}

func TestCredentialsFromMetadata(t *testing.T) {
	meta := Metadata{
		AccessKeyID:     "AKID",
		SecretAccessKey: "SECRET",
		SessionToken:    "TOKEN",
	}

	creds := CredentialsFromMetadata(meta)

	assert.Equal(t, "AKID", creds.AccessKeyID)
	assert.Equal(t, "SECRET", creds.SecretAccessKey)
	assert.Equal(t, "TOKEN", creds.SessionToken)
}

func TestParseDuration_Valid(t *testing.T) {
	d := ParseDuration("1h30m")
	assert.Equal(t, 90*time.Minute, d)
}

func TestParseDuration_Empty(t *testing.T) {
	d := ParseDuration("")
	assert.Equal(t, time.Duration(0), d)
}

func TestParseDuration_Invalid(t *testing.T) {
	d := ParseDuration("notaduration")
	assert.Equal(t, time.Duration(0), d)
}
