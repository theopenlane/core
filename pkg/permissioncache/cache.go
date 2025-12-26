package permissioncache

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/theopenlane/common/models"
)

const (
	featureCacheKeyPrefix  = "features"
	roleCacheKeyPrefix     = "role"
	userOrgsCacheKeyPrefix = "user_orgs"

	defaultTTL = 5 * time.Minute
)

// Cache stores enabled feature sets in Redis keyed by organization ID.  It
// reuses the same Redis instance as the session store but keeps feature data
// separate from user session values so entries can expire independently.
type Cache struct {
	Redis *redis.Client
	TTL   time.Duration
}

// NewCache returns a Cache using the provided redis client.
func NewCache(r *redis.Client, opts ...CacheOptions) *Cache {
	cache := &Cache{Redis: r, TTL: defaultTTL}

	for _, opt := range opts {
		opt(cache)
	}

	return cache
}

// CacheOptions is a functional option type for configuring the Cache.
type CacheOptions func(*Cache)

// WithCacheTTL sets the time-to-live for cache entries. If ttl is zero, the default TTL will be used.
func WithCacheTTL(ttl time.Duration) CacheOptions {
	return func(c *Cache) {
		if ttl > 0 {
			c.TTL = ttl
		}
	}
}

// GetFeatures retrieves the feature list for an organization.
func (c *Cache) GetFeatures(ctx context.Context, orgID string) ([]models.OrgModule, error) {
	strs, err := c.get(ctx, c.featureKey(orgID))
	if err != nil {
		return nil, err
	}

	if strs == nil {
		return nil, nil
	}

	features := make([]models.OrgModule, 0, len(strs))
	for _, s := range strs {
		features = append(features, models.OrgModule(s))
	}

	return features, nil
}

// GetRoles retrieves the role for a subject ID.
func (c *Cache) GetRoles(ctx context.Context, subjectID, orgID string) ([]string, error) {
	return c.get(ctx, c.roleKey(subjectID, orgID))
}

// HasRole checks if a role exists for a subject ID.
func (c *Cache) HasRole(ctx context.Context, subjectID, orgID string, role string) (bool, error) {
	return c.contains(ctx, c.roleKey(subjectID, orgID), role)
}

// GetUserOrgs retrieves the organizations a user belongs to.
func (c *Cache) GetUserOrgs(ctx context.Context, subjectID string) ([]string, error) {
	return c.get(ctx, c.userOrgsKey(subjectID))
}

func (c *Cache) HasOrgAccess(ctx context.Context, subjectID, orgID string) (bool, error) {
	return c.contains(ctx, c.userOrgsKey(subjectID), orgID)
}

// SetFeatures stores the feature list for an organization.
func (c *Cache) SetFeatures(ctx context.Context, orgID string, values []models.OrgModule) error {
	features := make([]string, 0, len(values))
	for _, feature := range values {
		features = append(features, feature.String())
	}

	return c.set(ctx, c.featureKey(orgID), features)
}

// SetRole stores the role for a subject ID.
func (c *Cache) SetRole(ctx context.Context, subjectID, orgID string, value string) error {
	roles, err := c.get(ctx, c.roleKey(subjectID, orgID))
	if err != nil {
		return err
	}

	// If the roles slice is nil, it means there are no roles set yet for this subjectID and orgID
	if roles == nil {
		roles = []string{}
	} else if slices.Contains(roles, value) {
		// If the role already exists, we do not need to add it again
		return nil
	}

	// append the new role to the existing roles
	roles = append(roles, value)

	return c.set(ctx, c.roleKey(subjectID, orgID), roles)
}

func (c *Cache) featureKey(orgID string) string {
	return fmt.Sprintf("%s:%s", featureCacheKeyPrefix, orgID)
}

func (c *Cache) roleKey(subjectID, orgID string) string {
	return fmt.Sprintf("%s:%s:%s", roleCacheKeyPrefix, subjectID, orgID)
}

func (c *Cache) userOrgsKey(subjectID string) string {
	// This key is used to store the organizations a user belongs to
	return fmt.Sprintf("%s:%s", userOrgsCacheKeyPrefix, subjectID)
}

// get returns the cached permissions for the given reference id and key or nil when absent.
func (c *Cache) get(ctx context.Context, key string) ([]string, error) {
	if c == nil || c.Redis == nil {
		return nil, nil
	}

	vals, err := c.Redis.SMembers(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}

	return vals, err
}

func (c *Cache) contains(ctx context.Context, key string, value string) (bool, error) {
	if c == nil || c.Redis == nil {
		return false, nil
	}

	exists, err := c.Redis.SIsMember(ctx, key, value).Result()
	if err != nil {
		return false, err
	}

	return exists, nil
}

// set stores the given values in the cache under the specified key.
func (c *Cache) set(ctx context.Context, key string, values []string) error {
	if c == nil || c.Redis == nil {
		return nil
	}

	pipe := c.Redis.TxPipeline()

	pipe.Del(ctx, key)

	if len(values) > 0 {
		pipe.SAdd(ctx, key, values)
		pipe.Expire(ctx, key, c.TTL)
	}

	_, err := pipe.Exec(ctx)

	return err
}
