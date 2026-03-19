package keymaker

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/theopenlane/core/internal/integrations/types"
)

const defaultAuthStateStoreMaxEntries = 4096

// AuthState captures the temporary state required to complete a definition auth flow callback
type AuthState struct {
	// State is the unique CSRF token identifying this authorization session
	State string
	// DefinitionID identifies which definition is handling the authorization
	DefinitionID string
	// InstallationID identifies the installation record being activated
	InstallationID string
	// CredentialRef identifies which credential slot the auth result should persist into.
	CredentialRef types.CredentialRef
	// CallbackState holds the opaque state payload returned by the definition's AuthStartFunc
	CallbackState json.RawMessage
	// CreatedAt records when the session was initiated
	CreatedAt time.Time
	// ExpiresAt specifies when the session becomes invalid
	ExpiresAt time.Time
}

// AuthStateStore persists callback state until the definition auth callback is completed.
// This is intentionally ephemeral and scoped to the definition auth flow lifecycle.
type AuthStateStore interface {
	Save(state AuthState) error
	Take(token string) (AuthState, error)
}

// InMemoryAuthStateStore stores definition auth callback state in process memory and is safe for concurrent use
type InMemoryAuthStateStore struct {
	// mu protects concurrent access to the sessions map
	mu sync.Mutex
	// sessions indexes authorization state by state token
	sessions map[string]AuthState
	// maxEntries bounds in-memory session growth under abandoned callback flows
	maxEntries int
	// now provides the current timestamp, overridable in tests
	now func() time.Time
}

// NewInMemoryAuthStateStore returns an in-memory definition authorization state store
func NewInMemoryAuthStateStore() *InMemoryAuthStateStore {
	return &InMemoryAuthStateStore{
		sessions:   map[string]AuthState{},
		maxEntries: defaultAuthStateStoreMaxEntries,
		now:        time.Now,
	}
}

// Save records the provided definition authorization state
func (m *InMemoryAuthStateStore) Save(state AuthState) error {
	if state.State == "" {
		return ErrAuthStateTokenRequired
	}

	clone := state
	now := m.now()
	if clone.CreatedAt.IsZero() {
		clone.CreatedAt = now
	}

	if clone.ExpiresAt.IsZero() {
		clone.ExpiresAt = clone.CreatedAt.Add(defaultSessionTTL)
	}

	token := clone.State

	m.mu.Lock()
	defer m.mu.Unlock()

	m.purgeExpiredLocked(now)

	if _, exists := m.sessions[token]; !exists && len(m.sessions) >= m.maxEntries {
		return ErrAuthStateStoreFull
	}

	m.sessions[token] = clone

	return nil
}

// Take retrieves and deletes authorization state associated with the given token
func (m *InMemoryAuthStateStore) Take(token string) (AuthState, error) {
	if token == "" {
		return AuthState{}, ErrAuthStateTokenRequired
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	now := m.now()

	session, ok := m.sessions[token]
	if !ok {
		m.purgeExpiredLocked(now)
		return AuthState{}, ErrAuthStateNotFound
	}

	if !session.ExpiresAt.IsZero() && now.After(session.ExpiresAt) {
		delete(m.sessions, token)
		m.purgeExpiredLocked(now)
		return AuthState{}, ErrAuthStateExpired
	}

	delete(m.sessions, token)
	m.purgeExpiredLocked(now)

	return session, nil
}

func (m *InMemoryAuthStateStore) purgeExpiredLocked(now time.Time) {
	for key, session := range m.sessions {
		if !session.ExpiresAt.IsZero() && !session.ExpiresAt.After(now) {
			delete(m.sessions, key)
		}
	}
}
