package keystore

import (
	"context"
	"sync"
	"time"

	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/pkg/integrations/registry"
)

// cacheSkew defines how far before token expiry the cache entry should be invalidated
const cacheSkew = 30 * time.Second

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

	// now returns the current time, overridable for testing
	now func() time.Time
}

// cacheKey uniquely identifies a cached credential entry
type cacheKey struct {
	// orgID identifies the organization owning the credential
	orgID string
	// provider identifies which provider issued the credential
	provider types.ProviderType
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
		store:    store,
		registry: reg,
		cache:    make(map[cacheKey]cachedCredential),
		now:      time.Now,
	}
}

// Get returns the latest credential payload for the given org/provider pair (using cache when valid)
func (b *Broker) Get(ctx context.Context, orgID string, provider types.ProviderType) (types.CredentialPayload, error) {
	if payload, ok := b.getCached(orgID, provider); ok {
		return payload, nil
	}

	payload, err := b.store.LoadCredential(ctx, orgID, provider)
	if err != nil {
		return types.CredentialPayload{}, err
	}

	b.setCached(orgID, provider, payload)

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

	b.setCached(orgID, provider, persisted)

	return persisted, nil
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

// getCached retrieves a cached credential if it exists and is not expired
func (b *Broker) getCached(orgID string, provider types.ProviderType) (types.CredentialPayload, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	entry, ok := b.cache[cacheKey{orgID: orgID, provider: provider}]
	if !ok {
		return types.CredentialPayload{}, false
	}

	if entry.expires.Before(b.now()) {
		return types.CredentialPayload{}, false
	}

	return entry.payload, true
}

// setCached stores the credential payload in the cache with an expiry time
func (b *Broker) setCached(orgID string, provider types.ProviderType, payload types.CredentialPayload) {
	b.mu.Lock()
	defer b.mu.Unlock()

	expiry := cacheExpiry(payload, b.now)

	b.cache[cacheKey{orgID: orgID, provider: provider}] = cachedCredential{
		payload: payload,
		expires: expiry,
	}
}

// cacheExpiry determines the cache expiry time based on the payload's token expiry
func cacheExpiry(payload types.CredentialPayload, now func() time.Time) time.Time {
	if payload.Token != nil && !payload.Token.Expiry.IsZero() {
		expires := payload.Token.Expiry.Add(-cacheSkew)
		if expires.After(now()) {
			return expires
		}
	}

	return now().Add(cacheSkew)
}
