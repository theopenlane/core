package interceptors_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/pkg/permissioncache"
	"github.com/theopenlane/core/pkg/testutils"
)

func ctxWithFeatures(org string, feats []string) context.Context {
	ctx := context.Background()
	r := testutils.NewRedisClient()
	cache := permissioncache.NewCache(r, permissioncache.WithCacheTTL(time.Minute))
	_ = cache.SetFeatures(ctx, org, feats)
	ctx = permissioncache.WithCache(ctx, cache)
	ctx = auth.WithAuthenticatedUser(ctx, &auth.AuthenticatedUser{OrganizationID: org})
	return ctx
}

func TestInterceptorRequireAnyFeature(t *testing.T) {
	ctx := ctxWithFeatures("org", []string{"a"})

	itc := interceptors.InterceptorRequireAnyFeature("a")
	fn := itc.(intercept.TraverseFunc)
	err := fn(ctx, nil)
	require.NoError(t, err)

	itc = interceptors.InterceptorRequireAnyFeature("b")
	fn = itc.(intercept.TraverseFunc)
	err = fn(ctx, nil)
	assert.Equal(t, interceptors.ErrFeatureNotEnabled, err)
}
