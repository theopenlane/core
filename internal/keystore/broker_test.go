package keystore

import (
	"testing"
	"time"

	"github.com/theopenlane/core/internal/integrations/types"
)

func TestBrokerGetCachedPurgesExpiredEntries(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	provider := types.ProviderType("acme")

	expiredKey := cacheKey{orgID: "org-1", provider: provider}
	validKey := cacheKey{orgID: "org-2", provider: provider}

	broker := &Broker{
		cache: map[cacheKey]cachedCredential{
			expiredKey: {
				credential: credentialWithExpiry(now.Add(-time.Minute), "expired"),
				expires:    now.Add(-time.Second),
			},
			validKey: {
				credential: credentialWithExpiry(now.Add(time.Hour), "valid"),
				expires:    now.Add(time.Minute),
			},
		},
		maxCacheEntries: 10,
		now:             func() time.Time { return now },
	}

	payload, ok := broker.getCached(validKey.orgID, provider, "")
	if !ok {
		t.Fatalf("expected valid cache hit")
	}
	if payload.credential.APIToken != "valid" {
		t.Fatalf("expected valid cached payload, got %q", payload.credential.APIToken)
	}

	if len(broker.cache) != 1 {
		t.Fatalf("expected expired entry to be purged, cache size=%d", len(broker.cache))
	}
	if _, exists := broker.cache[expiredKey]; exists {
		t.Fatalf("expected expired cache entry to be removed")
	}
}

func TestBrokerSetCachedEvictsOldestWhenCapacityReached(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	provider := types.ProviderType("acme")

	broker := &Broker{
		cache:           map[cacheKey]cachedCredential{},
		maxCacheEntries: 2,
		now:             func() time.Time { return now },
	}

	broker.setCached("org-1", provider, "", credentialSnapshot{credential: credentialWithExpiry(now.Add(5*time.Minute), "one")})
	broker.setCached("org-2", provider, "", credentialSnapshot{credential: credentialWithExpiry(now.Add(10*time.Minute), "two")})
	broker.setCached("org-3", provider, "", credentialSnapshot{credential: credentialWithExpiry(now.Add(15*time.Minute), "three")})

	if len(broker.cache) != 2 {
		t.Fatalf("expected cache size to remain bounded at 2, got %d", len(broker.cache))
	}

	if _, exists := broker.cache[cacheKey{orgID: "org-1", provider: provider}]; exists {
		t.Fatalf("expected oldest cache entry to be evicted")
	}
	if _, exists := broker.cache[cacheKey{orgID: "org-2", provider: provider}]; !exists {
		t.Fatalf("expected second cache entry to remain")
	}
	if _, exists := broker.cache[cacheKey{orgID: "org-3", provider: provider}]; !exists {
		t.Fatalf("expected third cache entry to be stored")
	}
}

func credentialWithExpiry(expiry time.Time, token string) types.CredentialSet {
	payload := types.CredentialSet{
		APIToken:         token,
		OAuthAccessToken: token,
	}
	payload.OAuthExpiry = &expiry
	return payload
}
