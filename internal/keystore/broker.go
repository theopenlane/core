package keystore

import (
	"context"
	"sync"
	"time"

	"github.com/theopenlane/core/internal/integrations/types"
)

// cacheSkew defines how far before token expiry the cache entry should be invalidated
const cacheSkew = 30 * time.Second

// nonExpiringCredentialTTL is the cache duration for credentials that carry no token expiry
const nonExpiringCredentialTTL = 5 * time.Minute

const defaultBrokerCacheMaxEntries = 4096

// DefinitionResolver looks up integration definitions for auth refresh
type DefinitionResolver interface {
	Definition(id types.DefinitionID) (types.Definition, bool)
}

// Broker caches and refreshes credentials for v2 installation records
type Broker struct {
	// store persists and retrieves credential sets
	store *Store
	// definitions resolves definition instances by ID for auth refresh
	definitions DefinitionResolver
	// mu protects concurrent access to the cache
	mu sync.Mutex
	// cache stores recently used credentials to avoid database roundtrips
	cache map[string]cachedCredential
	// maxCacheEntries bounds the number of in-memory cached credential entries
	maxCacheEntries int
	// now returns the current time, overridable for testing
	now func() time.Time
}

// cachedCredential holds a credential set and its cache expiry
type cachedCredential struct {
	// credential contains cached credential fields
	credential types.CredentialSet
	// expires specifies when this cache entry should be invalidated
	expires time.Time
}

// NewBroker constructs a Broker backed by the supplied store and definition resolver
func NewBroker(store *Store, definitions DefinitionResolver) (*Broker, error) {
	if store == nil || definitions == nil {
		return nil, ErrBrokerNotInitialized
	}

	return &Broker{
		store:           store,
		definitions:     definitions,
		cache:           make(map[string]cachedCredential),
		maxCacheEntries: defaultBrokerCacheMaxEntries,
		now:             time.Now,
	}, nil
}

// Get returns the cached or persisted credential for the given installation.
// Reports (zero, false, nil) when no credential has been stored yet.
func (b *Broker) Get(ctx context.Context, installationID string) (types.CredentialSet, bool, error) {
	if installationID == "" {
		return types.CredentialSet{}, false, ErrInstallationIDRequired
	}

	if cs, ok := b.getCached(installationID); ok {
		return cs, true, nil
	}

	integration, err := b.store.db.Integration.Get(integrationV2SystemContext(ctx), installationID)
	if err != nil {
		return types.CredentialSet{}, false, err
	}

	cs, ok, err := b.store.LoadCredential(ctx, integration)
	if err != nil {
		return types.CredentialSet{}, false, err
	}

	if ok {
		b.setCached(installationID, cs)
	}

	return cs, ok, nil
}

// Mint refreshes the credential for the given installation via the definition's auth refresh function
// and persists the result. Returns the refreshed credential.
func (b *Broker) Mint(ctx context.Context, installationID string) (types.CredentialSet, error) {
	if installationID == "" {
		return types.CredentialSet{}, ErrInstallationIDRequired
	}

	sysCtx := integrationV2SystemContext(ctx)

	integration, err := b.store.db.Integration.Get(sysCtx, installationID)
	if err != nil {
		return types.CredentialSet{}, err
	}

	defID := types.DefinitionID(integration.DefinitionID)
	def, ok := b.definitions.Definition(defID)
	if !ok {
		return types.CredentialSet{}, ErrDefinitionNotFound
	}

	if def.Auth == nil || def.Auth.Refresh == nil {
		return types.CredentialSet{}, ErrRefreshNotSupported
	}

	current, _, err := b.store.LoadCredential(ctx, integration)
	if err != nil {
		return types.CredentialSet{}, err
	}

	refreshed, err := def.Auth.Refresh(ctx, current)
	if err != nil {
		return types.CredentialSet{}, err
	}

	if err := b.store.SaveCredential(ctx, integration, refreshed); err != nil {
		return types.CredentialSet{}, err
	}

	b.setCached(installationID, refreshed)

	return refreshed, nil
}

// getCached retrieves a cached credential if it exists and has not expired
func (b *Broker) getCached(installationID string) (types.CredentialSet, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.purgeExpiredLocked(b.now())

	entry, ok := b.cache[installationID]
	if !ok {
		return types.CredentialSet{}, false
	}

	return entry.credential, true
}

// setCached stores the credential in the cache with an expiry time
func (b *Broker) setCached(installationID string, cs types.CredentialSet) {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := b.now()
	b.purgeExpiredLocked(now)

	if _, exists := b.cache[installationID]; !exists && len(b.cache) >= b.maxCacheEntries {
		b.evictOldestLocked()
	}

	b.cache[installationID] = cachedCredential{
		credential: cs,
		expires:    cacheExpiry(cs, b.now),
	}
}

func (b *Broker) purgeExpiredLocked(now time.Time) {
	for key, entry := range b.cache {
		if !entry.expires.After(now) {
			delete(b.cache, key)
		}
	}
}

func (b *Broker) evictOldestLocked() {
	var (
		oldestKey    string
		oldestExpiry time.Time
		found        bool
	)

	for key, entry := range b.cache {
		if !found || entry.expires.Before(oldestExpiry) {
			oldestKey = key
			oldestExpiry = entry.expires
			found = true
		}
	}

	if found {
		delete(b.cache, oldestKey)
	}
}

// cacheExpiry determines the cache expiry time for a credential set.
// Credentials with an OAuth expiry expire slightly before the token does (cacheSkew).
// Credentials without an expiry (API keys, service account metadata) use a longer TTL.
func cacheExpiry(credential types.CredentialSet, now func() time.Time) time.Time {
	if credential.OAuthExpiry != nil && !credential.OAuthExpiry.IsZero() {
		expires := credential.OAuthExpiry.Add(-cacheSkew)
		if expires.After(now()) {
			return expires
		}
	}

	return now().Add(nonExpiringCredentialTTL)
}
