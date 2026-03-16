package awskit

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMetadataFromProviderData_ValidData verifies provider data decoding preserves explicit values
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

// TestMetadataFromProviderData_DefaultSessionName verifies the fallback session name is applied
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

// TestMetadataFromProviderData_SessionNameFromData verifies provider data can override the session name
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

// TestMetadataFromProviderData_RegionFallback verifies the home region is reused when region is omitted
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

// TestMetadataFromProviderData_DefaultAccountScope verifies account scope defaults to all
func TestMetadataFromProviderData_DefaultAccountScope(t *testing.T) {
	raw, err := json.Marshal(map[string]any{
		"region": "us-east-1",
	})
	require.NoError(t, err)

	meta, err := MetadataFromProviderData(raw, "session")
	require.NoError(t, err)

	assert.Equal(t, AccountScopeAll, meta.AccountScope)
}

// TestMetadataFromProviderData_EmptyInput verifies empty provider data still yields defaults
func TestMetadataFromProviderData_EmptyInput(t *testing.T) {
	meta, err := MetadataFromProviderData(nil, "session")
	require.NoError(t, err)

	assert.Equal(t, AccountScopeAll, meta.AccountScope)
	assert.Equal(t, "session", meta.SessionName)
}

// TestCredentialsFromMetadata verifies AWS credentials are copied from metadata
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

// TestParseDuration_Valid verifies valid durations are parsed
func TestParseDuration_Valid(t *testing.T) {
	d := ParseDuration("1h30m")
	assert.Equal(t, 90*time.Minute, d)
}

// TestParseDuration_Empty verifies empty durations return zero
func TestParseDuration_Empty(t *testing.T) {
	d := ParseDuration("")
	assert.Equal(t, time.Duration(0), d)
}

// TestParseDuration_Invalid verifies invalid durations return zero
func TestParseDuration_Invalid(t *testing.T) {
	d := ParseDuration("notaduration")
	assert.Equal(t, time.Duration(0), d)
}
