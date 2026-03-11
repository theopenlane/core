package awssts

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/integrations/types"
)

func TestAwsSTSMetadataFromPayload_MissingProviderData(t *testing.T) {
	_, err := awsSTSMetadataFromPayload(types.CredentialSet{})
	require.ErrorIs(t, err, ErrProviderMetadataRequired)
}

func TestAwsSTSMetadataFromPayload_ValidData(t *testing.T) {
	raw, err := json.Marshal(map[string]any{
		"roleArn":         "  arn:aws:iam::123456789012:role/MyRole  ",
		"region":          "us-east-1",
		"accountId":       "123456789012",
		"accessKeyId":     "AKIAIOSFODNN7EXAMPLE",
		"secretAccessKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		"sessionToken":    "AQoDYXdzEJr",
	})
	require.NoError(t, err)

	meta, err := awsSTSMetadataFromPayload(types.CredentialSet{ProviderData: raw})
	require.NoError(t, err)

	assert.Equal(t, "arn:aws:iam::123456789012:role/MyRole", meta.RoleARN.String())
	assert.Equal(t, "us-east-1", meta.Region.String())
	assert.Equal(t, "123456789012", meta.AccountID.String())
	assert.Equal(t, "AKIAIOSFODNN7EXAMPLE", meta.AccessKeyID)
	assert.Equal(t, "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", meta.SecretAccessKey)
	assert.Equal(t, "AQoDYXdzEJr", meta.SessionToken)
}

func TestAwsSTSMetadataApplyDefaults_AccountScope(t *testing.T) {
	raw, err := json.Marshal(map[string]any{
		"roleArn": "arn:aws:iam::123:role/R",
		"region":  "us-west-2",
	})
	require.NoError(t, err)

	meta, err := awsSTSMetadataFromPayload(types.CredentialSet{ProviderData: raw})
	require.NoError(t, err)

	assert.Equal(t, AWSAccountScopeAll, meta.AccountScope.String())
}

func TestAwsSTSMetadataApplyDefaults_RegionHomeRegionFallback(t *testing.T) {
	raw, err := json.Marshal(map[string]any{
		"homeRegion": "eu-west-1",
	})
	require.NoError(t, err)

	meta, err := awsSTSMetadataFromPayload(types.CredentialSet{ProviderData: raw})
	require.NoError(t, err)

	assert.Equal(t, "eu-west-1", meta.Region.String())
	assert.Equal(t, "eu-west-1", meta.HomeRegion.String())
}

func TestMint_ReturnsProviderData(t *testing.T) {
	raw, err := json.Marshal(map[string]any{
		"roleArn":         "arn:aws:iam::123456789012:role/MyRole",
		"region":          "us-east-1",
		"accountId":       "123456789012",
		"accessKeyId":     "AKIAIOSFODNN7EXAMPLE",
		"secretAccessKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		"sessionToken":    "AQoDYXdzEJr",
	})
	require.NoError(t, err)

	p := &Provider{}

	result, err := p.Mint(context.Background(), types.CredentialMintRequest{
		Credential: types.CredentialSet{ProviderData: raw},
	})
	require.NoError(t, err)
	require.NotEmpty(t, result.ProviderData)

	var decoded combinedProviderData
	require.NoError(t, json.Unmarshal(result.ProviderData, &decoded))

	assert.Equal(t, "arn:aws:iam::123456789012:role/MyRole", decoded.RoleARN)
	assert.Equal(t, "us-east-1", decoded.Region)
	assert.Equal(t, "AKIAIOSFODNN7EXAMPLE", decoded.AccessKeyID)
	assert.Equal(t, "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", decoded.SecretAccessKey)
	assert.Equal(t, "AQoDYXdzEJr", decoded.SessionToken)
}

func TestMint_MissingProviderData(t *testing.T) {
	p := &Provider{}

	_, err := p.Mint(context.Background(), types.CredentialMintRequest{
		Credential: types.CredentialSet{},
	})
	require.ErrorIs(t, err, ErrProviderMetadataRequired)
}

func TestBeginAuth_NotSupported(t *testing.T) {
	p := &Provider{}

	_, err := p.BeginAuth(context.Background(), types.AuthContext{})
	require.ErrorIs(t, err, ErrBeginAuthNotSupported)
}
