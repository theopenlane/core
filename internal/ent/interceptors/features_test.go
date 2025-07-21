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

	// Add "_module" suffix to features to match the expected format in privacy rules
	moduleFeats := make([]string, len(feats))
	for i, feat := range feats {
		moduleFeats[i] = feat + "_module"
	}

	_ = cache.SetFeatures(ctx, org, moduleFeats)
	ctx = permissioncache.WithCache(ctx, cache)
	ctx = auth.WithAuthenticatedUser(ctx, &auth.AuthenticatedUser{OrganizationID: org})
	return ctx
}

func TestInterceptorRequireAnyFeature(t *testing.T) {
	ctx := ctxWithFeatures("org", []string{"a"})

	itc := interceptors.InterceptorRequireAnyFeature("test", "a")
	fn := itc.(intercept.TraverseFunc)
	err := fn(ctx, nil)
	require.NoError(t, err)

	itc = interceptors.InterceptorRequireAnyFeature("test", "b")
	fn = itc.(intercept.TraverseFunc)
	err = fn(ctx, nil)
	assert.Equal(t, interceptors.ErrFeatureNotEnabled, err)
}

func TestInterceptorRequireAllFeatures(t *testing.T) {
	ctx := ctxWithFeatures("org2", []string{"alpha", "beta", "gamma"})

	// test all features enabled
	itc := interceptors.InterceptorRequireAllFeatures("test", "alpha", "beta")
	fn := itc.(intercept.TraverseFunc)
	err := fn(ctx, nil)
	require.NoError(t, err)

	// test when not all the requested features are enabled
	itc = interceptors.InterceptorRequireAllFeatures("test", "alpha", "delta")
	fn = itc.(intercept.TraverseFunc)
	err = fn(ctx, nil)
	assert.Equal(t, interceptors.ErrFeatureNotEnabled, err)

	// single feature that exists
	itc = interceptors.InterceptorRequireAllFeatures("test", "gamma")
	fn = itc.(intercept.TraverseFunc)
	err = fn(ctx, nil)
	require.NoError(t, err)

	// single feature that doesn't exist
	itc = interceptors.InterceptorRequireAllFeatures("test", "omega")
	fn = itc.(intercept.TraverseFunc)
	err = fn(ctx, nil)
	assert.Equal(t, interceptors.ErrFeatureNotEnabled, err)

	// all enabled features
	itc = interceptors.InterceptorRequireAllFeatures("test", "alpha", "beta", "gamma")
	fn = itc.(intercept.TraverseFunc)
	err = fn(ctx, nil)
	require.NoError(t, err)
}
