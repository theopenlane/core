package features_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/pkg/features"
)

func TestContext(t *testing.T) {
	c := &features.Cache{}
	ctx := features.WithCache(context.Background(), c)
	got, ok := features.CacheFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, c, got)
}
