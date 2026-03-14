package keystore

import (
	"testing"
	"time"

	"github.com/theopenlane/core/internal/integrations/types"
)

func TestBrokerGetCachedPurgesExpiredEntries(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	broker := &Broker{
		cache: map[string]cachedCredential{
			"install-expired": {
				credential: credentialWithExpiry(now.Add(-time.Minute), "expired"),
				expires:    now.Add(-time.Second),
			},
			"install-valid": {
				credential: credentialWithExpiry(now.Add(time.Hour), "valid"),
				expires:    now.Add(time.Minute),
			},
		},
		maxCacheEntries: 10,
		now:             func() time.Time { return now },
	}

	cs, ok := broker.getCached("install-valid")
	if !ok {
		t.Fatalf("expected valid cache hit")
	}
	if cs.OAuthAccessToken != "valid" {
		t.Fatalf("expected valid cached payload, got %q", cs.OAuthAccessToken)
	}

	if len(broker.cache) != 1 {
		t.Fatalf("expected expired entry to be purged, cache size=%d", len(broker.cache))
	}
	if _, exists := broker.cache["install-expired"]; exists {
		t.Fatalf("expected expired cache entry to be removed")
	}
}

func TestBrokerSetCachedEvictsOldestWhenCapacityReached(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	broker := &Broker{
		cache:           map[string]cachedCredential{},
		maxCacheEntries: 2,
		now:             func() time.Time { return now },
	}

	broker.setCached("install-1", credentialWithExpiry(now.Add(5*time.Minute), "one"))
	broker.setCached("install-2", credentialWithExpiry(now.Add(10*time.Minute), "two"))
	broker.setCached("install-3", credentialWithExpiry(now.Add(15*time.Minute), "three"))

	if len(broker.cache) != 2 {
		t.Fatalf("expected cache size to remain bounded at 2, got %d", len(broker.cache))
	}

	if _, exists := broker.cache["install-1"]; exists {
		t.Fatalf("expected oldest cache entry to be evicted")
	}
	if _, exists := broker.cache["install-2"]; !exists {
		t.Fatalf("expected second cache entry to remain")
	}
	if _, exists := broker.cache["install-3"]; !exists {
		t.Fatalf("expected third cache entry to be stored")
	}
}

func credentialWithExpiry(expiry time.Time, token string) types.CredentialSet {
	cs := types.CredentialSet{
		OAuthAccessToken: token,
	}
	cs.OAuthExpiry = &expiry
	return cs
}
