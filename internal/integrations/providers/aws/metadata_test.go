package aws

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/integrations/types"
)

func TestAwsMetadataFromCredential_MissingProviderData(t *testing.T) {
	_, err := awsMetadataFromCredential(types.CredentialSet{}, awsDefaultSession)
	require.ErrorIs(t, err, ErrMetadataMissing)
}

func TestAwsMetadataFromCredential_MissingRoleARN(t *testing.T) {
	raw, err := json.Marshal(map[string]any{
		"region": "us-east-1",
	})
	require.NoError(t, err)

	_, err = awsMetadataFromCredential(types.CredentialSet{ProviderData: raw}, awsDefaultSession)
	require.ErrorIs(t, err, ErrRoleARNMissing)
}

func TestAwsMetadataFromCredential_MissingRegion(t *testing.T) {
	raw, err := json.Marshal(map[string]any{
		"roleArn": "arn:aws:iam::123:role/R",
	})
	require.NoError(t, err)

	_, err = awsMetadataFromCredential(types.CredentialSet{ProviderData: raw}, awsDefaultSession)
	require.ErrorIs(t, err, ErrRegionMissing)
}

func TestAwsMetadataFromCredential_Valid(t *testing.T) {
	raw, err := json.Marshal(map[string]any{
		"roleArn":   "arn:aws:iam::123456789012:role/MyRole",
		"region":    "us-east-1",
		"accountId": "123456789012",
	})
	require.NoError(t, err)

	meta, err := awsMetadataFromCredential(types.CredentialSet{ProviderData: raw}, awsDefaultSession)
	require.NoError(t, err)

	assert.Equal(t, "arn:aws:iam::123456789012:role/MyRole", meta.RoleARN)
	assert.Equal(t, "us-east-1", meta.Region)
	assert.Equal(t, "123456789012", meta.AccountID)
	assert.Equal(t, awsDefaultSession, meta.SessionName)
}
