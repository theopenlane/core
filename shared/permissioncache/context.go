package permissioncache

import (
	"context"

	"github.com/redis/go-redis/v9"
	echo "github.com/theopenlane/echox"
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

// SetCacheContext sets the cache context in the echo context, defaults to a ttl of 5 minutes if not specified
func SetCacheContext(c echo.Context, redisClient *redis.Client, opts ...CacheOptions) {
	ctx := c.Request().Context()

	ctx = WithCache(ctx, NewCache(redisClient, opts...))

	c.SetRequest(c.Request().WithContext(ctx))
}
