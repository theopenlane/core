package vulnintel

import (
	"context"
	"encoding/json"
	"time"

	"github.com/theopenlane/logx"

	"github.com/theopenlane/core/v2/internal/integrations/definitions/objectstore"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

const (
	// cacheKeyPrefix is the bucket prefix advisories are cached under, one object per advisory id
	cacheKeyPrefix = "vulnintel/osv/"
	// cacheTTL is how long a cached advisory is trusted before it is looked up again
	cacheTTL = 7 * 24 * time.Hour
	// missingCacheTTL is how long an advisory OSV does not know is remembered as missing
	missingCacheTTL = 24 * time.Hour
)

// cacheEntry is one cached advisory lookup, including lookups OSV had no advisory for
type cacheEntry struct {
	// FetchedAt is when the advisory was looked up
	FetchedAt time.Time `json:"fetched_at"`
	// Found reports whether OSV had the advisory
	Found bool `json:"found"`
	// Record is the advisory when found
	Record osvRecord `json:"record,omitzero"`
}

// fresh reports whether the entry is still within its TTL
func (e cacheEntry) fresh(now time.Time) bool {
	ttl := cacheTTL
	if !e.Found {
		ttl = missingCacheTTL
	}

	return now.Sub(e.FetchedAt) < ttl
}

// advisoryCache keeps OSV lookups in the objectstore runtime bucket, shared by every replica and organization
// since OSV data is public, and turns itself off for the run when the objectstore runtime is unavailable
type advisoryCache struct {
	// services runs the objectstore runtime operations
	services types.RuntimeServices
	// disabled is set once a cache operation fails so the run stops retrying it
	disabled bool
}

// get returns the cached lookup for the advisory id when one is present and fresh
func (c *advisoryCache) get(ctx context.Context, id string, now time.Time) (cacheEntry, bool) {
	if c.disabled {
		return cacheEntry{}, false
	}

	config, err := json.Marshal(objectstore.ReadObject{Key: cacheKeyPrefix + id + ".json"})
	if err != nil {
		return cacheEntry{}, false
	}

	response, err := c.services.ExecuteRuntimeOperation(ctx, objectstore.DefinitionID.ID(), objectstore.ReadObjectOp.Name(), config)
	if err != nil {
		c.disable(ctx, err)

		return cacheEntry{}, false
	}

	var result objectstore.ReadObjectResult
	if err := json.Unmarshal(response, &result); err != nil || !result.Found {
		return cacheEntry{}, false
	}

	var entry cacheEntry
	if err := json.Unmarshal(result.Content, &entry); err != nil || !entry.fresh(now) {
		return cacheEntry{}, false
	}

	return entry, true
}

// put stores a lookup for the advisory id
func (c *advisoryCache) put(ctx context.Context, id string, entry cacheEntry) {
	if c.disabled {
		return
	}

	config, err := json.Marshal(objectstore.WriteObject{Key: cacheKeyPrefix + id + ".json", Content: entry})
	if err != nil {
		return
	}

	if _, err := c.services.ExecuteRuntimeOperation(ctx, objectstore.DefinitionID.ID(), objectstore.WriteObjectOp.Name(), config); err != nil {
		c.disable(ctx, err)
	}
}

// disable turns the cache off for the rest of the run
func (c *advisoryCache) disable(ctx context.Context, err error) {
	logx.FromContext(ctx).Warn().Err(err).Msg("vulnintel: objectstore cache unavailable, looking advisories up directly")

	c.disabled = true
}
