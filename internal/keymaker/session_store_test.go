package keymaker

import (
	"context"
	"errors"
	"testing"

	"github.com/theopenlane/core/internal/integrations"
	"github.com/theopenlane/core/internal/integrations/types"
)

func TestMemorySessionStoreSaveAndTake(t *testing.T) {
	t.Parallel()

	store := NewMemorySessionStore()
	session := ActivationSession{
		State:       "state-1",
		AuthSession: &fakeAuthSession{state: "state-1", provider: types.ProviderType("acme")},
	}

	if err := store.Save(context.Background(), session); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	retrieved, err := store.Take(context.Background(), "state-1")
	if err != nil {
		t.Fatalf("Take error: %v", err)
	}

	if retrieved.State != session.State {
		t.Fatalf("expected state %s, got %s", session.State, retrieved.State)
	}

	_, err = store.Take(context.Background(), "state-1")
	if !errors.Is(err, integrations.ErrAuthorizationStateNotFound) {
		t.Fatalf("expected ErrAuthorizationStateNotFound, got %v", err)
	}
}

func TestMemorySessionStoreSaveValidatesInput(t *testing.T) {
	t.Parallel()

	store := NewMemorySessionStore()

	if err := store.Save(context.Background(), ActivationSession{}); !errors.Is(err, integrations.ErrStateRequired) {
		t.Fatalf("expected ErrStateRequired, got %v", err)
	}

	err := store.Save(context.Background(), ActivationSession{State: "state"})
	if !errors.Is(err, integrations.ErrAuthSessionInvalid) {
		t.Fatalf("expected ErrAuthSessionInvalid, got %v", err)
	}
}

func TestMemorySessionStoreTakeValidatesState(t *testing.T) {
	t.Parallel()

	store := NewMemorySessionStore()
	_, err := store.Take(context.Background(), "")
	if !errors.Is(err, integrations.ErrStateRequired) {
		t.Fatalf("expected ErrStateRequired, got %v", err)
	}
}
