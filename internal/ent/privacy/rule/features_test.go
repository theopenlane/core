package rule_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/pkg/catalog/features"
	"github.com/theopenlane/core/pkg/testutils"
)

func setupContext(org string, feats []string) context.Context {
	ctx := context.Background()
	r := testutils.NewRedisClient()
	cache := features.NewCache(r, time.Minute)
	_ = cache.Set(ctx, org, feats)
	ctx = features.WithCache(ctx, cache)
	ctx = auth.WithAuthenticatedUser(ctx, &auth.AuthenticatedUser{OrganizationID: org})
	return ctx
}

func TestHasFeature(t *testing.T) {
	ctx := setupContext("org1", []string{"alpha", "beta"})

	ok, err := rule.HasFeature(ctx, "alpha")
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = rule.HasFeature(ctx, "gamma")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestHasAnyFeature(t *testing.T) {
	ctx := setupContext("org2", []string{"foo", "bar"})

	ok, err := rule.HasAnyFeature(ctx, "baz", "bar")
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = rule.HasAnyFeature(ctx, "baz", "qux")
	require.NoError(t, err)
	assert.False(t, ok)
}
