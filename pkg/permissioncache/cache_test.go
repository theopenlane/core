package permissioncache_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/pkg/permissioncache"
	"github.com/theopenlane/core/pkg/testutils"
)

func TestCacheSetGet(t *testing.T) {
	ctx := context.Background()
	r := testutils.NewRedisClient()
	c := permissioncache.NewCache(r, permissioncache.WithCacheTTL(time.Minute))

	features := []models.OrgModule{models.OrgModule("a"), models.OrgModule("b")}
	err := c.SetFeatures(ctx, "org1", features)
	assert.NoError(t, err)

	vals, err := c.GetFeatures(ctx, "org1")
	assert.NoError(t, err)
	assert.ElementsMatch(t, features, vals)
}

func TestCacheNil(t *testing.T) {
	ctx := context.Background()
	var c *permissioncache.Cache

	vals, err := c.GetFeatures(ctx, "org")
	assert.NoError(t, err)
	assert.Nil(t, vals)

	features := []models.OrgModule{models.OrgModule("x")}
	err = c.SetFeatures(ctx, "org", features)
	assert.NoError(t, err)
}

func TestNewCacheDefaultTTL(t *testing.T) {
	c := permissioncache.NewCache(nil)
	assert.Equal(t, 5*time.Minute, c.TTL)
}
