package keymaker

import (
	"errors"
	"testing"
	"time"

	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/internal/integrations"
)

func TestMemorySessionStoreSaveAndTake(t *testing.T) {
	t.Parallel()

	store := NewMemorySessionStore()
	session := ActivationSession{
		State:       "state-1",
		AuthSession: &fakeAuthSession{state: "state-1", provider: types.ProviderType("acme")},
	}

	if err := store.Save(session); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	retrieved, err := store.Take("state-1")
	if err != nil {
		t.Fatalf("Take error: %v", err)
	}

	if retrieved.State != session.State {
		t.Fatalf("expected state %s, got %s", session.State, retrieved.State)
	}

	_, err = store.Take("state-1")
	if !errors.Is(err, integrations.ErrAuthorizationStateNotFound) {
		t.Fatalf("expected ErrAuthorizationStateNotFound, got %v", err)
	}
}

func TestMemorySessionStoreSaveValidatesInput(t *testing.T) {
	t.Parallel()

	store := NewMemorySessionStore()

	if err := store.Save(ActivationSession{}); !errors.Is(err, integrations.ErrStateRequired) {
		t.Fatalf("expected ErrStateRequired, got %v", err)
	}

	err := store.Save(ActivationSession{State: "state"})
	if !errors.Is(err, integrations.ErrAuthSessionInvalid) {
		t.Fatalf("expected ErrAuthSessionInvalid, got %v", err)
	}
}

func TestMemorySessionStoreTakeValidatesState(t *testing.T) {
	t.Parallel()

	store := NewMemorySessionStore()
	_, err := store.Take("")
	if !errors.Is(err, integrations.ErrStateRequired) {
		t.Fatalf("expected ErrStateRequired, got %v", err)
	}
}

func TestMemorySessionStoreTakeReturnsExpired(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	store := NewMemorySessionStore()
	store.maxEntries = 10
	store.now = func() time.Time { return now }

	session := ActivationSession{
		State:       "state-expired",
		AuthSession: &fakeAuthSession{state: "state-expired", provider: types.ProviderType("acme")},
		CreatedAt:   now.Add(-2 * time.Minute),
		ExpiresAt:   now.Add(-time.Minute),
	}

	if err := store.Save(session); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	_, err := store.Take("state-expired")
	if !errors.Is(err, integrations.ErrAuthorizationStateExpired) {
		t.Fatalf("expected ErrAuthorizationStateExpired, got %v", err)
	}
}

func TestMemorySessionStoreSaveEnforcesCapacity(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	store := NewMemorySessionStore()
	store.maxEntries = 1
	store.now = func() time.Time { return now }

	first := ActivationSession{
		State:       "state-1",
		AuthSession: &fakeAuthSession{state: "state-1", provider: types.ProviderType("acme")},
		CreatedAt:   now,
		ExpiresAt:   now.Add(time.Minute),
	}
	if err := store.Save(first); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	second := ActivationSession{
		State:       "state-2",
		AuthSession: &fakeAuthSession{state: "state-2", provider: types.ProviderType("acme")},
		CreatedAt:   now,
		ExpiresAt:   now.Add(time.Minute),
	}
	err := store.Save(second)
	if !errors.Is(err, integrations.ErrAuthorizationStateStoreFull) {
		t.Fatalf("expected ErrAuthorizationStateStoreFull, got %v", err)
	}
}
