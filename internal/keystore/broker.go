package keystore

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
)

// ensure Broker satisfies IntegrationCredentialSource at compile time
var _ IntegrationCredentialSource = (*Broker)(nil)

// cacheSkew defines how far before token expiry the cache entry should be invalidated
const cacheSkew = 30 * time.Second

// nonExpiringCredentialTTL is the cache duration for credentials that carry no token expiry,
// such as API keys and service account metadata; long enough to avoid frequent DB reads
const nonExpiringCredentialTTL = 5 * time.Minute

const defaultBrokerCacheMaxEntries = 4096

// Broker exchanges persisted credentials for short-lived tokens via registered providers
type Broker struct {
	// store persists and retrieves credential payloads
	store *Store
	// registry provides access to provider implementations for minting
	registry *registry.Registry

	// mu protects concurrent access to the cache
	mu sync.RWMutex
	// cache stores recently used credentials to avoid database roundtrips
	cache map[cacheKey]cachedCredential
	// maxCacheEntries bounds the number of in-memory cached credential entries
	maxCacheEntries int

	// now returns the current time, overridable for testing
	now func() time.Time
}

// cacheKey uniquely identifies a cached credential entry
type cacheKey struct {
	// orgID identifies the organization owning the credential
	orgID string
	// provider identifies which provider issued the credential
	provider types.ProviderType
	// integrationID scopes cache entries to a specific installed integration when provided
	integrationID string
}

// cachedCredential holds a credential payload and its expiry time
type cachedCredential struct {
	// payload contains the cached credential data
	payload types.CredentialPayload
	// expires specifies when this cache entry should be invalidated
	expires time.Time
}

// NewBroker constructs a broker backed by the supplied store and provider registry
func NewBroker(store *Store, reg *registry.Registry) *Broker {
	return &Broker{
		store:           store,
		registry:        reg,
		cache:           make(map[cacheKey]cachedCredential),
		maxCacheEntries: defaultBrokerCacheMaxEntries,
		now:             time.Now,
	}
}

// Get returns the latest credential payload for the given org/provider pair (using cache when valid)
func (b *Broker) Get(ctx context.Context, orgID string, provider types.ProviderType) (types.CredentialPayload, error) {
	if payload, ok := b.getCached(orgID, provider, ""); ok {
		return payload, nil
	}

	payload, err := b.store.LoadCredential(ctx, orgID, provider)
	if err != nil {
		return types.CredentialPayload{}, err
	}

	b.setCached(orgID, provider, "", payload)

	return payload, nil
}

// GetForIntegration returns credentials scoped to a specific integration record
func (b *Broker) GetForIntegration(ctx context.Context, orgID string, provider types.ProviderType, integrationID string) (types.CredentialPayload, error) {
	if payload, ok := b.getCached(orgID, provider, integrationID); ok {
		return payload, nil
	}

	providerInstance, err := b.lookupProvider(provider)
	if err != nil {
		return types.CredentialPayload{}, err
	}

	if providerInstance.Capabilities().EnvironmentCredentials {
		return b.MintForIntegration(ctx, orgID, provider, integrationID)
	}

	payload, err := b.store.LoadCredentialForIntegration(ctx, orgID, provider, integrationID)
	if err != nil {
		return types.CredentialPayload{}, err
	}

	b.setCached(orgID, provider, integrationID, payload)

	return payload, nil
}

// Mint refreshes the stored credential via the provider and returns the updated payload
func (b *Broker) Mint(ctx context.Context, orgID string, provider types.ProviderType) (types.CredentialPayload, error) {
	providerInstance, err := b.lookupProvider(provider)
	if err != nil {
		return types.CredentialPayload{}, err
	}

	stored, err := b.store.LoadCredential(ctx, orgID, provider)
	if err != nil {
		return types.CredentialPayload{}, err
	}

	subject := types.CredentialSubject{
		Provider:   provider,
		OrgID:      orgID,
		Credential: stored,
	}

	minted, err := providerInstance.Mint(ctx, subject)
	if err != nil {
		return types.CredentialPayload{}, err
	}

	if minted.Provider == types.ProviderUnknown {
		minted.Provider = provider
	}

	persisted, err := b.store.SaveCredential(ctx, orgID, minted)
	if err != nil {
		return types.CredentialPayload{}, err
	}

	b.setCached(orgID, provider, "", persisted)

	return persisted, nil
}

// MintForIntegration refreshes and persists credentials scoped to a specific integration record
func (b *Broker) MintForIntegration(ctx context.Context, orgID string, provider types.ProviderType, integrationID string) (types.CredentialPayload, error) {
	providerInstance, err := b.lookupProvider(provider)
	if err != nil {
		return types.CredentialPayload{}, err
	}

	stored, persistedCredential, err := b.loadIntegrationSubject(ctx, orgID, provider, integrationID)
	if err != nil {
		return types.CredentialPayload{}, err
	}

	subject := types.CredentialSubject{
		Provider:      provider,
		OrgID:         orgID,
		IntegrationID: integrationID,
		Credential:    stored,
	}

	minted, err := providerInstance.Mint(ctx, subject)
	if err != nil {
		return types.CredentialPayload{}, err
	}

	if minted.Provider == types.ProviderUnknown {
		minted.Provider = provider
	}

	if minted.ProviderState == nil && stored.ProviderState != nil {
		minted.ProviderState = stored.ProviderState
	}

	if providerInstance.Capabilities().EnvironmentCredentials || !persistedCredential {
		b.setCached(orgID, provider, integrationID, minted)

		return minted, nil
	}

	persisted, err := b.store.SaveCredentialForIntegration(ctx, orgID, integrationID, minted)
	if err != nil {
		return types.CredentialPayload{}, err
	}

	b.setCached(orgID, provider, integrationID, persisted)

	return persisted, nil
}

// loadIntegrationSubject loads the persisted credential for a specific integration record.
// Returns the payload, a boolean indicating whether a persisted credential was found, and any error.
// When the credential is not found and the provider uses environment credentials, falls back to
// loading the integration subject record for minting
func (b *Broker) loadIntegrationSubject(ctx context.Context, orgID string, provider types.ProviderType, integrationID string) (types.CredentialPayload, bool, error) {
	stored, err := b.store.LoadCredentialForIntegration(ctx, orgID, provider, integrationID)
	if err == nil {
		return stored, true, nil
	}

	if !errors.Is(err, ErrCredentialNotFound) {
		return types.CredentialPayload{}, false, err
	}

	// Credential not found — check if provider uses environment-level credentials
	// and fall back to loading the integration subject for minting
	providerInstance, lookupErr := b.lookupProvider(provider)
	if lookupErr != nil || !providerInstance.Capabilities().EnvironmentCredentials {
		return types.CredentialPayload{}, false, err
	}

	return b.store.loadCredentialSubjectForIntegration(ctx, orgID, provider, integrationID)
}

// lookupProvider retrieves the provider instance from the registry
func (b *Broker) lookupProvider(provider types.ProviderType) (types.Provider, error) {
	if b.registry == nil {
		return nil, ErrProviderNotRegistered
	}

	instance, ok := b.registry.Provider(provider)
	if !ok {
		return nil, ErrProviderNotRegistered
	}

	return instance, nil
}

// getCached retrieves a cached credential if it exists and has not expired.
// Uses a read lock so concurrent credential lookups do not serialize.
// Expired entries are not returned but are not actively purged here; purge happens on write
func (b *Broker) getCached(orgID string, provider types.ProviderType, integrationID string) (types.CredentialPayload, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	entry, ok := b.cache[cacheKey{orgID: orgID, provider: provider, integrationID: integrationID}]
	if !ok {
		return types.CredentialPayload{}, false
	}

	if !entry.expires.After(b.now()) {
		return types.CredentialPayload{}, false
	}

	return entry.payload, true
}

// setCached stores the credential payload in the cache with an expiry time
func (b *Broker) setCached(orgID string, provider types.ProviderType, integrationID string, payload types.CredentialPayload) {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := b.now()
	b.purgeExpiredLocked(now)

	key := cacheKey{orgID: orgID, provider: provider, integrationID: integrationID}
	if _, exists := b.cache[key]; !exists && len(b.cache) >= b.maxCacheEntries {
		b.evictOldestLocked()
	}

	expiry := cacheExpiry(payload, b.now)
	b.cache[key] = cachedCredential{
		payload: payload,
		expires: expiry,
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
		oldestKey    cacheKey
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

// cacheExpiry determines the cache expiry time based on the payload's token expiry.
// For credentials with a token expiry, the entry expires slightly before the token does (cacheSkew).
// For credentials without an expiry (API keys, service account metadata), a longer TTL is used
// to avoid frequent database reads
func cacheExpiry(payload types.CredentialPayload, now func() time.Time) time.Time {
	if payload.Token != nil && !payload.Token.Expiry.IsZero() {
		expires := payload.Token.Expiry.Add(-cacheSkew)
		if expires.After(now()) {
			return expires
		}
	}

	return now().Add(nonExpiringCredentialTTL)
}
