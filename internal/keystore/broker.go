package keystore

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// ensure Broker satisfies IntegrationCredentialSource at compile time.
var _ IntegrationCredentialSource = (*Broker)(nil)

// cacheSkew defines how far before token expiry the cache entry should be invalidated.
const cacheSkew = 30 * time.Second

// nonExpiringCredentialTTL is the cache duration for credentials that carry no token expiry,
// such as API keys and service account metadata; long enough to avoid frequent DB reads.
const nonExpiringCredentialTTL = 5 * time.Minute

const defaultBrokerCacheMaxEntries = 4096

// Broker exchanges persisted credentials for short-lived tokens via registered providers.
type Broker struct {
	// store persists and retrieves credential sets.
	store *Store
	// registry provides access to provider implementations for minting.
	registry *registry.Registry
	// mu protects concurrent access to the cache.
	mu sync.RWMutex
	// cache stores recently used credentials to avoid database roundtrips.
	cache map[cacheKey]cachedCredential
	// maxCacheEntries bounds the number of in-memory cached credential entries.
	maxCacheEntries int
	// now returns the current time, overridable for testing.
	now func() time.Time
}

// cacheKey uniquely identifies a cached credential entry.
type cacheKey struct {
	// orgID identifies the organization owning the credential.
	orgID string
	// provider identifies which provider issued the credential.
	provider types.ProviderType
	// integrationID scopes cache entries to a specific installed integration when provided.
	integrationID string
}

// cachedCredential holds credential data, metadata, and cache expiry.
type cachedCredential struct {
	// credential contains cached credential fields.
	credential types.CredentialSet
	// authKind stores persisted auth kind metadata.
	authKind types.AuthKind
	// providerState stores integration provider state used by environment credentials.
	providerState *types.IntegrationProviderState
	// expires specifies when this cache entry should be invalidated.
	expires time.Time
}

// credentialSnapshot carries credential data with metadata for mint/persist decisions.
type credentialSnapshot struct {
	credential    types.CredentialSet
	authKind      types.AuthKind
	providerState *types.IntegrationProviderState
}

// NewBroker constructs a broker backed by the supplied store and provider registry.
func NewBroker(store *Store, reg *registry.Registry) (*Broker, error) {
	if store == nil || reg == nil {
		return nil, ErrBrokerNotInitialized
	}

	return &Broker{
		store:           store,
		registry:        reg,
		cache:           make(map[cacheKey]cachedCredential),
		maxCacheEntries: defaultBrokerCacheMaxEntries,
		now:             time.Now,
	}, nil
}

// Get returns the latest credential set for the given org/provider pair (using cache when valid).
func (b *Broker) Get(ctx context.Context, orgID string, provider types.ProviderType) (types.CredentialSet, error) {
	if snapshot, ok := b.getCached(orgID, provider, ""); ok {
		return snapshot.credential, nil
	}

	credential, authKind, providerState, err := b.store.LoadCredential(ctx, orgID, provider)
	if err != nil {
		return types.CredentialSet{}, err
	}

	b.setCached(orgID, provider, "", credentialSnapshot{
		credential:    credential,
		authKind:      authKind,
		providerState: providerState,
	})

	return types.CloneCredentialSet(credential), nil
}

// GetForIntegration returns credentials scoped to a specific integration record.
func (b *Broker) GetForIntegration(ctx context.Context, orgID string, provider types.ProviderType, integrationID string) (types.CredentialSet, error) {
	if snapshot, ok := b.getCached(orgID, provider, integrationID); ok {
		return snapshot.credential, nil
	}

	providerInstance, err := b.lookupProvider(provider)
	if err != nil {
		return types.CredentialSet{}, err
	}

	if providerInstance.Capabilities().EnvironmentCredentials {
		return b.MintForIntegration(ctx, orgID, provider, integrationID)
	}

	credential, authKind, providerState, err := b.store.LoadCredentialForIntegration(ctx, orgID, provider, integrationID)
	if err != nil {
		return types.CredentialSet{}, err
	}

	b.setCached(orgID, provider, integrationID, credentialSnapshot{
		credential:    credential,
		authKind:      authKind,
		providerState: providerState,
	})

	return types.CloneCredentialSet(credential), nil
}

// Mint refreshes the stored credential via the provider and returns the updated credential set.
func (b *Broker) Mint(ctx context.Context, orgID string, provider types.ProviderType) (types.CredentialSet, error) {
	providerInstance, err := b.lookupProvider(provider)
	if err != nil {
		return types.CredentialSet{}, err
	}

	stored, err := b.loadProviderSnapshot(ctx, orgID, provider)
	if err != nil {
		return types.CredentialSet{}, err
	}

	minted, err := providerInstance.Mint(ctx, types.CredentialMintRequest{
		Provider:      provider,
		OrgID:         orgID,
		Credential:    types.CloneCredentialSet(stored.credential),
		ProviderState: cloneProviderState(stored.providerState),
	})
	if err != nil {
		return types.CredentialSet{}, err
	}

	resolvedKind := normalizeMintedAuthKind(stored.authKind, minted)

	persisted, err := b.store.SaveCredential(ctx, orgID, provider, resolvedKind, minted)
	if err != nil {
		return types.CredentialSet{}, err
	}

	b.setCached(orgID, provider, "", credentialSnapshot{
		credential:    persisted,
		authKind:      resolvedKind,
		providerState: stored.providerState,
	})

	return types.CloneCredentialSet(persisted), nil
}

// MintForIntegration refreshes and persists credentials scoped to a specific integration record.
func (b *Broker) MintForIntegration(ctx context.Context, orgID string, provider types.ProviderType, integrationID string) (types.CredentialSet, error) {
	providerInstance, err := b.lookupProvider(provider)
	if err != nil {
		return types.CredentialSet{}, err
	}

	stored, persistedCredential, err := b.loadIntegrationSubject(ctx, orgID, provider, integrationID)
	if err != nil {
		return types.CredentialSet{}, err
	}

	minted, err := providerInstance.Mint(ctx, types.CredentialMintRequest{
		Provider:      provider,
		OrgID:         orgID,
		IntegrationID: integrationID,
		Credential:    types.CloneCredentialSet(stored.credential),
		ProviderState: cloneProviderState(stored.providerState),
	})
	if err != nil {
		return types.CredentialSet{}, err
	}

	resolvedKind := normalizeMintedAuthKind(stored.authKind, minted)

	if providerInstance.Capabilities().EnvironmentCredentials || !persistedCredential {
		b.setCached(orgID, provider, integrationID, credentialSnapshot{
			credential:    minted,
			authKind:      resolvedKind,
			providerState: stored.providerState,
		})

		return types.CloneCredentialSet(minted), nil
	}

	persisted, err := b.store.SaveCredentialForIntegration(ctx, orgID, integrationID, provider, resolvedKind, minted)
	if err != nil {
		return types.CredentialSet{}, err
	}

	b.setCached(orgID, provider, integrationID, credentialSnapshot{
		credential:    persisted,
		authKind:      resolvedKind,
		providerState: stored.providerState,
	})

	return types.CloneCredentialSet(persisted), nil
}

func (b *Broker) loadProviderSnapshot(ctx context.Context, orgID string, provider types.ProviderType) (credentialSnapshot, error) {
	credential, authKind, providerState, err := b.store.LoadCredential(ctx, orgID, provider)
	if err != nil {
		return credentialSnapshot{}, err
	}

	return credentialSnapshot{
		credential:    credential,
		authKind:      authKind,
		providerState: providerState,
	}, nil
}

// loadIntegrationSubject loads credentials for a specific integration record.
// It returns persistedCredential=false when no persisted secret was found.
func (b *Broker) loadIntegrationSubject(ctx context.Context, orgID string, provider types.ProviderType, integrationID string) (credentialSnapshot, bool, error) {
	credential, authKind, providerState, err := b.store.LoadCredentialForIntegration(ctx, orgID, provider, integrationID)
	if err == nil {
		return credentialSnapshot{
			credential:    credential,
			authKind:      authKind,
			providerState: providerState,
		}, true, nil
	}

	if !errors.Is(err, ErrCredentialNotFound) {
		return credentialSnapshot{}, false, err
	}

	// Credential not found: check environment credentials and fall back to integration subject metadata.
	providerInstance, lookupErr := b.lookupProvider(provider)
	if lookupErr != nil || !providerInstance.Capabilities().EnvironmentCredentials {
		return credentialSnapshot{}, false, err
	}

	credential, authKind, providerState, found, loadErr := b.store.loadCredentialForIntegrationRecord(ctx, orgID, provider, integrationID)
	if loadErr != nil {
		return credentialSnapshot{}, false, loadErr
	}

	return credentialSnapshot{
		credential:    credential,
		authKind:      authKind,
		providerState: providerState,
	}, found, nil
}

// lookupProvider retrieves the provider instance from the registry.
func (b *Broker) lookupProvider(provider types.ProviderType) (types.Provider, error) {
	instance, ok := b.registry.Provider(provider)
	if !ok {
		return nil, ErrProviderNotRegistered
	}

	return instance, nil
}

// getCached retrieves a cached credential if it exists and has not expired.
// Expired entries are purged on read to keep cache size bounded when write traffic is low.
func (b *Broker) getCached(orgID string, provider types.ProviderType, integrationID string) (credentialSnapshot, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.purgeExpiredLocked(b.now())

	entry, ok := b.cache[cacheKey{orgID: orgID, provider: provider, integrationID: integrationID}]
	if !ok {
		return credentialSnapshot{}, false
	}

	return credentialSnapshot{
		credential:    types.CloneCredentialSet(entry.credential),
		authKind:      entry.authKind,
		providerState: cloneProviderState(entry.providerState),
	}, true
}

// setCached stores the credential snapshot in the cache with an expiry time.
func (b *Broker) setCached(orgID string, provider types.ProviderType, integrationID string, snapshot credentialSnapshot) {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := b.now()
	b.purgeExpiredLocked(now)

	key := cacheKey{orgID: orgID, provider: provider, integrationID: integrationID}
	if _, exists := b.cache[key]; !exists && len(b.cache) >= b.maxCacheEntries {
		b.evictOldestLocked()
	}

	expiry := cacheExpiry(snapshot.credential, b.now)
	b.cache[key] = cachedCredential{
		credential:    types.CloneCredentialSet(snapshot.credential),
		authKind:      snapshot.authKind,
		providerState: cloneProviderState(snapshot.providerState),
		expires:       expiry,
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

// cacheExpiry determines the cache expiry time based on OAuth expiry fields.
// For credentials with OAuth expiry, the entry expires slightly before the token does (cacheSkew).
// For credentials without expiry (API keys, service account metadata), a longer TTL is used.
func cacheExpiry(credential types.CredentialSet, now func() time.Time) time.Time {
	if credential.OAuthExpiry != nil && !credential.OAuthExpiry.IsZero() {
		expires := credential.OAuthExpiry.Add(-cacheSkew)
		if expires.After(now()) {
			return expires
		}
	}

	return now().Add(nonExpiringCredentialTTL)
}

func normalizeMintedAuthKind(storedKind types.AuthKind, credential types.CredentialSet) types.AuthKind {
	normalized := storedKind.Normalize()
	if normalized != types.AuthKindUnknown {
		return normalized
	}

	return types.InferAuthKind(credential)
}

func cloneProviderState(state *types.IntegrationProviderState) *types.IntegrationProviderState {
	if state == nil {
		return nil
	}

	cloned := types.IntegrationProviderState{}
	if len(state.Providers) > 0 {
		cloned.Providers = make(map[string]json.RawMessage, len(state.Providers))
		for key, value := range state.Providers {
			cloned.Providers[key] = jsonx.CloneRawMessage(value)
		}
	}

	return &cloned
}
