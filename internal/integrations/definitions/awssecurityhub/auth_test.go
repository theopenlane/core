package awssecurityhub

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCredentialSchemaFromProviderData_ValidData verifies provider data decoding preserves explicit values.
func TestCredentialSchemaFromProviderData_ValidData(t *testing.T) {
	raw, err := json.Marshal(map[string]any{
		"roleArn":         "arn:aws:iam::123456789012:role/MyRole",
		"homeRegion":      "us-east-1",
		"accountId":       "123456789012",
		"accountScope":    "all",
		"accessKeyId":     "AKIAIOSFODNN7EXAMPLE",
		"secretAccessKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		"sessionToken":    "AQoDYXdzEJr",
	})
	require.NoError(t, err)

	credential, err := credentialSchemaFromProviderData(raw)
	require.NoError(t, err)

	assert.Equal(t, "arn:aws:iam::123456789012:role/MyRole", credential.RoleARN)
	assert.Equal(t, "us-east-1", credential.HomeRegion)
	assert.Equal(t, "123456789012", credential.AccountID)
	assert.Equal(t, "all", credential.AccountScope)
	assert.Equal(t, "AKIAIOSFODNN7EXAMPLE", credential.AccessKeyID)
	assert.Equal(t, "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", credential.SecretAccessKey)
	assert.Equal(t, "AQoDYXdzEJr", credential.SessionToken)
}

// TestCredentialSchemaFromProviderData_DefaultSessionName verifies the default session name is applied.
func TestCredentialSchemaFromProviderData_DefaultSessionName(t *testing.T) {
	raw, err := json.Marshal(map[string]any{
		"roleArn":    "arn:aws:iam::123:role/R",
		"homeRegion": "us-west-2",
	})
	require.NoError(t, err)

	credential, err := credentialSchemaFromProviderData(raw)
	require.NoError(t, err)

	assert.Equal(t, defaultSessionName, credential.SessionName)
}

// TestCredentialSchemaFromProviderData_SessionNameFromData verifies provider data can override the session name.
func TestCredentialSchemaFromProviderData_SessionNameFromData(t *testing.T) {
	raw, err := json.Marshal(map[string]any{
		"roleArn":     "arn:aws:iam::123:role/R",
		"homeRegion":  "us-west-2",
		"sessionName": "custom-session",
	})
	require.NoError(t, err)

	credential, err := credentialSchemaFromProviderData(raw)
	require.NoError(t, err)

	assert.Equal(t, "custom-session", credential.SessionName)
}

// TestCredentialSchemaFromProviderData_DefaultAccountScope verifies account scope defaults to all.
func TestCredentialSchemaFromProviderData_DefaultAccountScope(t *testing.T) {
	raw, err := json.Marshal(map[string]any{
		"homeRegion": "us-east-1",
	})
	require.NoError(t, err)

	credential, err := credentialSchemaFromProviderData(raw)
	require.NoError(t, err)

	assert.Equal(t, AccountScopeAll, credential.AccountScope)
}

// TestCredentialSchemaFromProviderData_EmptyInput verifies empty provider data still yields defaults.
func TestCredentialSchemaFromProviderData_EmptyInput(t *testing.T) {
	credential, err := credentialSchemaFromProviderData(nil)
	require.NoError(t, err)

	assert.Equal(t, AccountScopeAll, credential.AccountScope)
	assert.Equal(t, defaultSessionName, credential.SessionName)
}

// TestParseDuration_Valid verifies valid durations are parsed.
func TestParseDuration_Valid(t *testing.T) {
	duration := parseDuration("1h30m")
	assert.Equal(t, 90*time.Minute, duration)
}

// TestParseDuration_Empty verifies empty durations return zero.
func TestParseDuration_Empty(t *testing.T) {
	duration := parseDuration("")
	assert.Equal(t, time.Duration(0), duration)
}

// TestParseDuration_Invalid verifies invalid durations return zero.
func TestParseDuration_Invalid(t *testing.T) {
	duration := parseDuration("notaduration")
	assert.Equal(t, time.Duration(0), duration)
}
