package awssecurityhub

import (
	"testing"
	"time"

	securityhubtypes "github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gtassert "gotest.tools/v3/assert"
)

func TestBuildFilters(t *testing.T) {
	t.Parallel()

	t.Run("account scope all skips account filters", func(t *testing.T) {
		t.Parallel()

		filters := buildFilters(CollectionScope{
			AccountScope: AccountScopeAll,
			AccountIDs:   []string{"123456789012", "234567890123"},
		}, nil)
		assert.Empty(t, filters.AwsAccountId)
	})

	t.Run("account scope specific with ids adds account filters", func(t *testing.T) {
		t.Parallel()

		filters := buildFilters(CollectionScope{
			AccountScope: AccountScopeSpecific,
			AccountIDs:   []string{"111111111111", "222222222222"},
		}, nil)
		require.Len(t, filters.AwsAccountId, 2)
		assert.Equal(t, "111111111111", *filters.AwsAccountId[0].Value)
		assert.Equal(t, securityhubtypes.StringFilterComparisonEquals, filters.AwsAccountId[0].Comparison)
		assert.Equal(t, "222222222222", *filters.AwsAccountId[1].Value)
		assert.Equal(t, securityhubtypes.StringFilterComparisonEquals, filters.AwsAccountId[1].Comparison)
	})

	t.Run("account scope specific with no ids falls back to account id", func(t *testing.T) {
		t.Parallel()

		filters := buildFilters(CollectionScope{
			AccountID:    "111111111111",
			AccountScope: AccountScopeSpecific,
		}, nil)
		require.Len(t, filters.AwsAccountId, 1)
		assert.Equal(t, "111111111111", *filters.AwsAccountId[0].Value)
	})

	t.Run("linked regions adds region filters", func(t *testing.T) {
		t.Parallel()

		filters := buildFilters(CollectionScope{
			LinkedRegions: []string{"us-west-2", "eu-west-1"},
		}, nil)
		require.Len(t, filters.Region, 2)
		assert.Equal(t, "us-west-2", *filters.Region[0].Value)
		assert.Equal(t, securityhubtypes.StringFilterComparisonEquals, filters.Region[0].Comparison)
		assert.Equal(t, "eu-west-1", *filters.Region[1].Value)
		assert.Equal(t, securityhubtypes.StringFilterComparisonEquals, filters.Region[1].Comparison)
	})

	t.Run("no linked regions skips region filters", func(t *testing.T) {
		t.Parallel()

		filters := buildFilters(CollectionScope{}, nil)
		assert.Empty(t, filters.Region)
	})

	t.Run("account scope specific and linked regions both applied", func(t *testing.T) {
		t.Parallel()

		filters := buildFilters(CollectionScope{
			AccountScope:  AccountScopeSpecific,
			AccountIDs:    []string{"111111111111"},
			LinkedRegions: []string{"us-west-2"},
		}, nil)
		require.Len(t, filters.AwsAccountId, 1)
		assert.Equal(t, "111111111111", *filters.AwsAccountId[0].Value)
		require.Len(t, filters.Region, 1)
		assert.Equal(t, "us-west-2", *filters.Region[0].Value)
	})

	t.Run("lastRunAt 90 days ago sets UpdatedAt filter to that window", func(t *testing.T) {
		t.Parallel()

		lastRunAt := time.Now().UTC().Add(-90 * 24 * time.Hour)

		filters := buildFilters(CollectionScope{}, &lastRunAt)
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

		before := time.Now().UTC().Truncate(time.Second)
		ts := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
		filters := buildFilters(CollectionScope{}, &ts)
		after := time.Now().UTC()

		require.Len(t, filters.UpdatedAt, 1)
		gtassert.Equal(t, "2025-01-15T12:00:00Z", *filters.UpdatedAt[0].Start)
		require.NotNil(t, filters.UpdatedAt[0].End)

		end, err := time.Parse(time.RFC3339, *filters.UpdatedAt[0].End)
		require.NoError(t, err)
		gtassert.Assert(t, !end.Before(before) && !end.After(after), "End should be within the call window [%v, %v], got %v", before, after, end)
	})
}
