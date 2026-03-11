package awssecurityhub

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/integrations/providers/awskit"
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

func TestSecurityHubFiltersFromMetadata_NoFilters(t *testing.T) {
	meta := awskit.Metadata{
		AccountScope: awskit.AccountScopeAll,
	}

	filters := securityHubFiltersFromMetadata(meta)
	assert.Nil(t, filters)
}

func TestSecurityHubFiltersFromMetadata_SpecificAccounts(t *testing.T) {
	meta := awskit.Metadata{
		AccountScope: awskit.AccountScopeSpecific,
		AccountIDs:   []string{"111111111111", "222222222222"},
	}

	filters := securityHubFiltersFromMetadata(meta)
	assert.NotNil(t, filters)
	assert.Len(t, filters.AwsAccountId, 2)
}

func TestSecurityHubFiltersFromMetadata_LinkedRegions(t *testing.T) {
	meta := awskit.Metadata{
		AccountScope:  awskit.AccountScopeAll,
		LinkedRegions: []string{"us-east-1", "eu-west-1"},
	}

	filters := securityHubFiltersFromMetadata(meta)
	assert.NotNil(t, filters)
	assert.Len(t, filters.Region, 2)
}

func TestBuilder_Type(t *testing.T) {
	b := Builder()
	assert.Equal(t, TypeAWSSecurityHub, b.Type())
}
