package awssecurityhub

import (
	"context"
	"encoding/json"
	"testing"

	securityhubtypes "github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/types"
)

func makeAssumeRoleBindings(t *testing.T, schema AssumeRoleCredentialSchema) types.CredentialBindings {
	t.Helper()

	data, err := json.Marshal(schema)
	require.NoError(t, err)

	return types.CredentialBindings{
		{
			Ref:        awsAssumeRoleCredential.ID(),
			Credential: models.CredentialSet{Data: json.RawMessage(data)},
		},
	}
}

func TestBuildFilters(t *testing.T) {
	t.Parallel()

	t.Run("missing credential returns error", func(t *testing.T) {
		t.Parallel()

		_, err := buildFilters(context.Background(), types.CredentialBindings{})
		require.ErrorIs(t, err, ErrCredentialMetadataRequired)
	})

	t.Run("account scope all skips account filters", func(t *testing.T) {
		t.Parallel()

		creds := makeAssumeRoleBindings(t, AssumeRoleCredentialSchema{
			RoleARN:      "arn:aws:iam::123456789012:role/test-role",
			HomeRegion:   "us-east-1",
			AccountScope: AccountScopeAll,
			AccountIDs:   []string{"123456789012", "234567890123"},
		})

		filters, err := buildFilters(context.Background(), creds)
		require.NoError(t, err)
		assert.Empty(t, filters.AwsAccountId)
	})

	t.Run("account scope specific with ids adds account filters", func(t *testing.T) {
		t.Parallel()

		creds := makeAssumeRoleBindings(t, AssumeRoleCredentialSchema{
			RoleARN:      "arn:aws:iam::123456789012:role/test-role",
			HomeRegion:   "us-east-1",
			AccountScope: AccountScopeSpecific,
			AccountIDs:   []string{"111111111111", "222222222222"},
		})

		filters, err := buildFilters(context.Background(), creds)
		require.NoError(t, err)
		require.Len(t, filters.AwsAccountId, 2)
		assert.Equal(t, "111111111111", *filters.AwsAccountId[0].Value)
		assert.Equal(t, securityhubtypes.StringFilterComparisonEquals, filters.AwsAccountId[0].Comparison)
		assert.Equal(t, "222222222222", *filters.AwsAccountId[1].Value)
		assert.Equal(t, securityhubtypes.StringFilterComparisonEquals, filters.AwsAccountId[1].Comparison)
	})

	t.Run("account scope specific with no ids skips account filters", func(t *testing.T) {
		t.Parallel()

		creds := makeAssumeRoleBindings(t, AssumeRoleCredentialSchema{
			RoleARN:      "arn:aws:iam::123456789012:role/test-role",
			HomeRegion:   "us-east-1",
			AccountID:    "111111111111",
			AccountScope: AccountScopeSpecific,
		})

		filters, err := buildFilters(context.Background(), creds)
		require.NoError(t, err)
		require.Len(t, filters.AwsAccountId, 1)
		assert.Equal(t, "111111111111", *filters.AwsAccountId[0].Value)
	})

	t.Run("linked regions adds region filters", func(t *testing.T) {
		t.Parallel()

		creds := makeAssumeRoleBindings(t, AssumeRoleCredentialSchema{
			RoleARN:       "arn:aws:iam::123456789012:role/test-role",
			HomeRegion:    "us-east-1",
			LinkedRegions: []string{"us-west-2", "eu-west-1"},
		})

		filters, err := buildFilters(context.Background(), creds)
		require.NoError(t, err)
		require.Len(t, filters.Region, 2)
		assert.Equal(t, "us-west-2", *filters.Region[0].Value)
		assert.Equal(t, securityhubtypes.StringFilterComparisonEquals, filters.Region[0].Comparison)
		assert.Equal(t, "eu-west-1", *filters.Region[1].Value)
		assert.Equal(t, securityhubtypes.StringFilterComparisonEquals, filters.Region[1].Comparison)
	})

	t.Run("no linked regions skips region filters", func(t *testing.T) {
		t.Parallel()

		creds := makeAssumeRoleBindings(t, AssumeRoleCredentialSchema{
			RoleARN:    "arn:aws:iam::123456789012:role/test-role",
			HomeRegion: "us-east-1",
		})

		filters, err := buildFilters(context.Background(), creds)
		require.NoError(t, err)
		assert.Empty(t, filters.Region)
	})

	t.Run("account scope specific and linked regions both applied", func(t *testing.T) {
		t.Parallel()

		creds := makeAssumeRoleBindings(t, AssumeRoleCredentialSchema{
			RoleARN:       "arn:aws:iam::123456789012:role/test-role",
			HomeRegion:    "us-east-1",
			AccountScope:  AccountScopeSpecific,
			AccountIDs:    []string{"111111111111"},
			LinkedRegions: []string{"us-west-2"},
		})

		filters, err := buildFilters(context.Background(), creds)
		require.NoError(t, err)
		require.Len(t, filters.AwsAccountId, 1)
		assert.Equal(t, "111111111111", *filters.AwsAccountId[0].Value)
		require.Len(t, filters.Region, 1)
		assert.Equal(t, "us-west-2", *filters.Region[0].Value)
	})
}
