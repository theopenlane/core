package awssecurityhub

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	securityhubtypes "github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gtassert "gotest.tools/v3/assert"

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

		_, err := buildFilters(context.Background(), types.CredentialBindings{}, nil)
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

		filters, err := buildFilters(context.Background(), creds, nil)
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

		filters, err := buildFilters(context.Background(), creds, nil)
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

		filters, err := buildFilters(context.Background(), creds, nil)
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

		filters, err := buildFilters(context.Background(), creds, nil)
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

		filters, err := buildFilters(context.Background(), creds, nil)
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

		filters, err := buildFilters(context.Background(), creds, nil)
		require.NoError(t, err)
		require.Len(t, filters.AwsAccountId, 1)
		assert.Equal(t, "111111111111", *filters.AwsAccountId[0].Value)
		require.Len(t, filters.Region, 1)
		assert.Equal(t, "us-west-2", *filters.Region[0].Value)
	})

	t.Run("lastRunAt 90 days ago sets UpdatedAt filter to that window", func(t *testing.T) {
		t.Parallel()

		creds := makeAssumeRoleBindings(t, AssumeRoleCredentialSchema{
			RoleARN:    "arn:aws:iam::123456789012:role/test-role",
			HomeRegion: "us-east-1",
		})

		lastRunAt := time.Now().UTC().Add(-90 * 24 * time.Hour)

		filters, err := buildFilters(context.Background(), creds, &lastRunAt)
		gtassert.NilError(t, err)
		gtassert.Assert(t, len(filters.UpdatedAt) == 1, "expected 1 UpdatedAt filter")

		start, err := time.Parse(time.RFC3339, *filters.UpdatedAt[0].Start)
		gtassert.NilError(t, err)

		diff := start.Sub(lastRunAt)
		if diff < 0 {
			diff = -diff
		}

		gtassert.Assert(t, diff < time.Second, "expected UpdatedAt.Start to match lastRunAt within 1s, got diff %v", diff)
	})

	t.Run("lastRunAt sets UpdatedAt filter", func(t *testing.T) {
		t.Parallel()

		creds := makeAssumeRoleBindings(t, AssumeRoleCredentialSchema{
			RoleARN:    "arn:aws:iam::123456789012:role/test-role",
			HomeRegion: "us-east-1",
		})

		before := time.Now().UTC().Truncate(time.Second)
		ts := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
		filters, err := buildFilters(context.Background(), creds, &ts)
		after := time.Now().UTC()

		require.NoError(t, err)
		require.Len(t, filters.UpdatedAt, 1)
		gtassert.Equal(t, "2025-01-15T12:00:00Z", *filters.UpdatedAt[0].Start)
		require.NotNil(t, filters.UpdatedAt[0].End)

		end, err := time.Parse(time.RFC3339, *filters.UpdatedAt[0].End)
		require.NoError(t, err)
		gtassert.Assert(t, !end.Before(before) && !end.After(after), "End should be within the call window [%v, %v], got %v", before, after, end)
	})
}
