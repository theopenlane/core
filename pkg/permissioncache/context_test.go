package permissioncache_test

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/pkg/permissioncache"
)

func TestContext(t *testing.T) {
	c := &permissioncache.Cache{}
	ctx := permissioncache.WithCache(context.Background(), c)
	got, ok := permissioncache.CacheFromContext(ctx)
	assert.Check(t, ok)
	assert.Check(t, is.Equal(c, got))
}
