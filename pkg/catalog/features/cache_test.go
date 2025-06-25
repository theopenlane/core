package features_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/pkg/catalog/features"
	"github.com/theopenlane/core/pkg/testutils"
)

func TestCacheSetGet(t *testing.T) {
	ctx := context.Background()
	r := testutils.NewRedisClient()
	c := features.NewCache(r, time.Minute)

	err := c.Set(ctx, "org1", []string{"a", "b"})
	require.NoError(t, err)

	vals, err := c.Get(ctx, "org1")
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"a", "b"}, vals)
}

func TestCacheNil(t *testing.T) {
	ctx := context.Background()
	var c *features.Cache

	vals, err := c.Get(ctx, "org")
	assert.NoError(t, err)
	assert.Nil(t, vals)

	err = c.Set(ctx, "org", []string{"x"})
	assert.NoError(t, err)
}

func TestNewCacheDefaultTTL(t *testing.T) {
	c := features.NewCache(nil, 0)
	assert.Equal(t, 5*time.Minute, c.TTL)
}
