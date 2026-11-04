package permissioncache_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/testutils"
	"github.com/theopenlane/core/v2/pkg/permissioncache"
)

func TestCacheSetGet(t *testing.T) {
	ctx := context.Background()
	r := testutils.NewRedisClient()
	c := permissioncache.NewCache(r, permissioncache.WithCacheTTL(time.Minute))

	features := []models.OrgModule{models.OrgModule("a"), models.OrgModule("b")}
	err := c.SetFeatures(ctx, "org1", features)
	assert.Check(t, err)

	vals, err := c.GetFeatures(ctx, "org1")
	assert.Check(t, err)
	assert.Check(t, is.DeepEqual(slices.Sorted(slices.Values(features)), slices.Sorted(slices.Values(vals))))
}

func TestCacheNil(t *testing.T) {
	ctx := context.Background()
	var c *permissioncache.Cache

	vals, err := c.GetFeatures(ctx, "org")
	assert.Check(t, err)
	assert.Check(t, is.Nil(vals))

	features := []models.OrgModule{models.OrgModule("x")}
	err = c.SetFeatures(ctx, "org", features)
	assert.Check(t, err)
}

func TestNewCacheDefaultTTL(t *testing.T) {
	c := permissioncache.NewCache(nil)
	assert.Check(t, is.Equal(5*time.Minute, c.TTL))
}
