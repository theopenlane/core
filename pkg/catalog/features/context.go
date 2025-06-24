package features

import (
	"context"

	"github.com/theopenlane/utils/contextx"
)

// WithCache stores the feature cache on the context for downstream use.
func WithCache(ctx context.Context, c *Cache) context.Context {
	return contextx.With(ctx, c)
}

// CacheFromContext retrieves the feature cache from context if present.
func CacheFromContext(ctx context.Context) (*Cache, bool) {
	return contextx.From[*Cache](ctx)
}
