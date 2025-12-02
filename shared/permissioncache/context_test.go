package permissioncache_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/shared/permissioncache"
)

func TestContext(t *testing.T) {
	c := &permissioncache.Cache{}
	ctx := permissioncache.WithCache(context.Background(), c)
	got, ok := permissioncache.CacheFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, c, got)
}
