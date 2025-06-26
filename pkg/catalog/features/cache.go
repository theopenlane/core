package features

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache stores enabled feature sets in Redis keyed by organization ID.  It
// reuses the same Redis instance as the session store but keeps feature data
// separate from user session values so entries can expire independently.
type Cache struct {
	Redis *redis.Client
	TTL   time.Duration
}

// NewCache returns a Cache using the provided redis client.
func NewCache(r *redis.Client, ttl time.Duration) *Cache {
	if ttl == 0 {
		ttl = 5 * time.Minute // nolint: mnd
	}

	return &Cache{Redis: r, TTL: ttl}
}

func (c *Cache) key(orgID string) string { return "features:" + orgID }

// Get returns the cached features for an organization or nil when absent.
func (c *Cache) Get(ctx context.Context, orgID string) ([]string, error) {
	if c == nil || c.Redis == nil {
		return nil, nil
	}

	vals, err := c.Redis.SMembers(ctx, c.key(orgID)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}

	return vals, err
}

// Set stores the feature list for an organization.
func (c *Cache) Set(ctx context.Context, orgID string, feats []string) error {
	if c == nil || c.Redis == nil {
		return nil
	}

	pipe := c.Redis.TxPipeline()

	pipe.Del(ctx, c.key(orgID))

	if len(feats) > 0 {
		pipe.SAdd(ctx, c.key(orgID), feats)
		pipe.Expire(ctx, c.key(orgID), c.TTL)
	}

	_, err := pipe.Exec(ctx)

	return err
}
