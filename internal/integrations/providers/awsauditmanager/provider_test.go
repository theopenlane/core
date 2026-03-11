package awsauditmanager

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/integrations/types"
)

func TestMetadataFromCredential_MissingProviderData(t *testing.T) {
	_, err := metadataFromCredential(types.CredentialSet{})
	require.ErrorIs(t, err, ErrMetadataMissing)
}

func TestMetadataFromCredential_MissingRoleARN(t *testing.T) {
	raw, err := json.Marshal(map[string]any{
		"region": "us-east-1",
	})
	require.NoError(t, err)

	_, err = metadataFromCredential(types.CredentialSet{ProviderData: raw})
	require.ErrorIs(t, err, ErrRoleARNMissing)
}

func TestMetadataFromCredential_MissingRegion(t *testing.T) {
	raw, err := json.Marshal(map[string]any{
		"roleArn": "arn:aws:iam::123:role/R",
	})
	require.NoError(t, err)

	_, err = metadataFromCredential(types.CredentialSet{ProviderData: raw})
	require.ErrorIs(t, err, ErrRegionMissing)
}

func TestMetadataFromCredential_Valid(t *testing.T) {
	raw, err := json.Marshal(map[string]any{
		"roleArn":   "arn:aws:iam::123456789012:role/MyRole",
		"region":    "us-east-1",
		"accountId": "123456789012",
	})
	require.NoError(t, err)

	meta, err := metadataFromCredential(types.CredentialSet{ProviderData: raw})
	require.NoError(t, err)

	assert.Equal(t, "arn:aws:iam::123456789012:role/MyRole", meta.RoleARN)
	assert.Equal(t, "us-east-1", meta.Region)
	assert.Equal(t, "123456789012", meta.AccountID)
}

func TestBuilder_Type(t *testing.T) {
	b := Builder()
	assert.Equal(t, TypeAWSAuditManager, b.Type())
}
