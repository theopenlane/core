package keymaker

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
)

func TestRedisAuthStateStoreSaveAndTake(t *testing.T) {
	t.Parallel()

	store := newRedisAuthStateStoreForTest(t)
	state := AuthState{
		State:          "token-1",
		DefinitionID:   "def_123",
		InstallationID: "int_123",
		CreatedAt:      time.Now().UTC(),
		ExpiresAt:      time.Now().UTC().Add(time.Minute),
	}

	if err := store.Save(state); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, err := store.Take(state.State)
	if err != nil {
		t.Fatalf("Take() error = %v", err)
	}

	if got.DefinitionID != state.DefinitionID {
		t.Fatalf("Take() definition_id = %q, want %q", got.DefinitionID, state.DefinitionID)
	}

	if _, err := store.Take(state.State); err != ErrAuthStateNotFound {
		t.Fatalf("second Take() error = %v, want %v", err, ErrAuthStateNotFound)
	}
}

func TestRedisAuthStateStoreRejectsExpiredSave(t *testing.T) {
	t.Parallel()

	store := newRedisAuthStateStoreForTest(t)
	now := time.Now().UTC()
	store.now = func() time.Time { return now }

	err := store.Save(AuthState{
		State:          "expired-token",
		DefinitionID:   "def_123",
		InstallationID: "int_123",
		CreatedAt:      now.Add(-2 * time.Minute),
		ExpiresAt:      now.Add(-time.Minute),
	})
	if err != ErrAuthStateExpired {
		t.Fatalf("Save() error = %v, want %v", err, ErrAuthStateExpired)
	}
}

func newRedisAuthStateStoreForTest(t *testing.T) *RedisAuthStateStore {
	t.Helper()

	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr:             server.Addr(),
		DisableIndentity: true, // compatibility with the pinned go-redis version
		MaintNotificationsConfig: &maintnotifications.Config{
			Mode: maintnotifications.ModeDisabled,
		},
	})

	t.Cleanup(func() {
		_ = client.Close()
		server.Close()
	})

	return NewRedisAuthStateStore(client)
}
