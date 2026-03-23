package awssecurityhub

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/integrations/types"
)

// TestResolveAssumeRoleCredential_ValidData verifies credential decoding preserves explicit values.
func TestResolveAssumeRoleCredential_ValidData(t *testing.T) {
	raw, err := json.Marshal(map[string]any{
		"roleArn":      "arn:aws:iam::123456789012:role/MyRole",
		"homeRegion":   "us-east-1",
		"accountId":    "123456789012",
		"accountScope": "all",
	})
	require.NoError(t, err)

	credential, err := resolveAssumeRoleCredential(types.CredentialBindings{
		{Ref: awsAssumeRoleCredential.ID(), Credential: types.CredentialSet{Data: raw}},
	})
	require.NoError(t, err)

	assert.Equal(t, "arn:aws:iam::123456789012:role/MyRole", credential.RoleARN)
	assert.Equal(t, "us-east-1", credential.HomeRegion)
	assert.Equal(t, "123456789012", credential.AccountID)
	assert.Equal(t, "all", credential.AccountScope)
}

// TestResolveAssumeRoleCredential_DefaultSessionName verifies the default session name is applied.
func TestResolveAssumeRoleCredential_DefaultSessionName(t *testing.T) {
	raw, err := json.Marshal(map[string]any{
		"roleArn":    "arn:aws:iam::123:role/R",
		"homeRegion": "us-west-2",
	})
	require.NoError(t, err)

	credential, err := resolveAssumeRoleCredential(types.CredentialBindings{
		{Ref: awsAssumeRoleCredential.ID(), Credential: types.CredentialSet{Data: raw}},
	})
	require.NoError(t, err)

	assert.Equal(t, defaultSessionName, credential.SessionName)
}

// TestResolveAssumeRoleCredential_SessionNameFromData verifies provider data can override the session name.
func TestResolveAssumeRoleCredential_SessionNameFromData(t *testing.T) {
	raw, err := json.Marshal(map[string]any{
		"roleArn":     "arn:aws:iam::123:role/R",
		"homeRegion":  "us-west-2",
		"sessionName": "custom-session",
	})
	require.NoError(t, err)

	credential, err := resolveAssumeRoleCredential(types.CredentialBindings{
		{Ref: awsAssumeRoleCredential.ID(), Credential: types.CredentialSet{Data: raw}},
	})
	require.NoError(t, err)

	assert.Equal(t, "custom-session", credential.SessionName)
}

// TestResolveAssumeRoleCredential_DefaultAccountScope verifies account scope defaults to all.
func TestResolveAssumeRoleCredential_DefaultAccountScope(t *testing.T) {
	raw, err := json.Marshal(map[string]any{
		"homeRegion": "us-east-1",
	})
	require.NoError(t, err)

	credential, err := resolveAssumeRoleCredential(types.CredentialBindings{
		{Ref: awsAssumeRoleCredential.ID(), Credential: types.CredentialSet{Data: raw}},
	})
	require.NoError(t, err)

	assert.Equal(t, AccountScopeAll, credential.AccountScope)
}

// TestResolveAssumeRoleCredential_EmptyInput verifies empty provider data is rejected.
func TestResolveAssumeRoleCredential_EmptyInput(t *testing.T) {
	_, err := resolveAssumeRoleCredential(types.CredentialBindings{
		{Ref: awsAssumeRoleCredential.ID(), Credential: types.CredentialSet{}},
	})
	require.ErrorIs(t, err, ErrCredentialMetadataRequired)
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
