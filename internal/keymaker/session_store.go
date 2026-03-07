package keymaker

import (
	"sync"
	"time"

	"github.com/theopenlane/core/internal/integrations"
	"github.com/theopenlane/core/internal/integrations/types"
)

const defaultAuthStateStoreMaxEntries = 4096

// AuthState captures the temporary state required to complete an OAuth flow callback.
type AuthState struct {
	// State is the unique CSRF token identifying this authorization session
	State string
	// Provider identifies which provider is handling the authorization
	Provider types.ProviderType
	// OrgID identifies the organization initiating the flow
	OrgID string
	// IntegrationID identifies the integration record being activated
	IntegrationID string
	// Scopes contains the authorization scopes requested from the provider
	Scopes []string
	// Metadata carries additional provider-specific configuration
	Metadata map[string]any
	// LabelOverrides customizes UI labels presented during authorization
	LabelOverrides map[string]string
	// CreatedAt records when the session was initiated
	CreatedAt time.Time
	// ExpiresAt specifies when the session becomes invalid
	ExpiresAt time.Time
	// AuthSession holds the provider-specific authorization state
	AuthSession types.AuthSession
}

// AuthStateStore persists callback state until the provider callback is completed.
// This is intentionally ephemeral and scoped to OAuth authorization state lifecycle.
type AuthStateStore interface {
	Save(state AuthState) error
	Take(token string) (AuthState, error)
}

// InMemoryAuthStateStore stores OAuth callback state in process memory and is safe for concurrent use
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

// NewInMemoryAuthStateStore returns an in-memory authorization state store
func NewInMemoryAuthStateStore() *InMemoryAuthStateStore {
	return &InMemoryAuthStateStore{
		sessions:   map[string]AuthState{},
		maxEntries: defaultAuthStateStoreMaxEntries,
		now:        time.Now,
	}
}

// Save records the provided authorization state.
func (m *InMemoryAuthStateStore) Save(state AuthState) error {
	if state.State == "" {
		return integrations.ErrStateRequired
	}

	if state.AuthSession == nil {
		return integrations.ErrAuthSessionInvalid
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
		return integrations.ErrAuthorizationStateStoreFull
	}

	clone.State = token
	m.sessions[token] = clone

	return nil
}

// Take retrieves and deletes authorization state associated with the given token
func (m *InMemoryAuthStateStore) Take(token string) (AuthState, error) {
	if token == "" {
		return AuthState{}, integrations.ErrStateRequired
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	now := m.now()
	session, ok := m.sessions[token]
	if !ok {
		m.purgeExpiredLocked(now)
		return AuthState{}, integrations.ErrAuthorizationStateNotFound
	}

	if !session.ExpiresAt.IsZero() && now.After(session.ExpiresAt) {
		delete(m.sessions, token)
		m.purgeExpiredLocked(now)
		return AuthState{}, integrations.ErrAuthorizationStateExpired
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
