package keymaker

import (
	"sync"
	"time"

	"github.com/theopenlane/core/internal/integrations"
	"github.com/theopenlane/core/internal/integrations/types"
)

const defaultMemorySessionStoreMaxEntries = 4096

// ActivationSession captures the temporary state required to complete an OAuth flow
type ActivationSession struct {
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

// SessionStore persists activation sessions until the provider callback is completed
type SessionStore interface {
	Save(session ActivationSession) error
	Take(state string) (ActivationSession, error)
}

// MemorySessionStore stores activation sessions in memory and is safe for concurrent use
type MemorySessionStore struct {
	// mu protects concurrent access to the sessions map
	mu sync.Mutex
	// sessions indexes activation sessions by their state token
	sessions map[string]ActivationSession
	// maxEntries bounds in-memory session growth under abandoned callback flows
	maxEntries int
	// now provides the current timestamp, overridable in tests
	now func() time.Time
}

// NewMemorySessionStore returns an in-memory session store
func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{
		sessions:   map[string]ActivationSession{},
		maxEntries: defaultMemorySessionStoreMaxEntries,
		now:        time.Now,
	}
}

// Save records the provided activation session
func (m *MemorySessionStore) Save(session ActivationSession) error {
	if session.State == "" {
		return integrations.ErrStateRequired
	}

	if session.AuthSession == nil {
		return integrations.ErrAuthSessionInvalid
	}

	clone := session
	now := m.now()
	if clone.CreatedAt.IsZero() {
		clone.CreatedAt = now
	}
	if clone.ExpiresAt.IsZero() {
		clone.ExpiresAt = clone.CreatedAt.Add(defaultSessionTTL)
	}
	state := clone.State

	m.mu.Lock()
	defer m.mu.Unlock()
	m.purgeExpiredLocked(now)
	if _, exists := m.sessions[state]; !exists && len(m.sessions) >= m.maxEntries {
		return integrations.ErrAuthorizationStateStoreFull
	}

	clone.State = state
	m.sessions[state] = clone

	return nil
}

// Take retrieves and deletes the session associated with the given state
func (m *MemorySessionStore) Take(state string) (ActivationSession, error) {
	if state == "" {
		return ActivationSession{}, integrations.ErrStateRequired
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	now := m.now()
	session, ok := m.sessions[state]
	if !ok {
		m.purgeExpiredLocked(now)
		return ActivationSession{}, integrations.ErrAuthorizationStateNotFound
	}

	if !session.ExpiresAt.IsZero() && now.After(session.ExpiresAt) {
		delete(m.sessions, state)
		m.purgeExpiredLocked(now)
		return ActivationSession{}, integrations.ErrAuthorizationStateExpired
	}

	delete(m.sessions, state)
	m.purgeExpiredLocked(now)

	return session, nil
}

func (m *MemorySessionStore) purgeExpiredLocked(now time.Time) {
	for key, session := range m.sessions {
		if !session.ExpiresAt.IsZero() && !session.ExpiresAt.After(now) {
			delete(m.sessions, key)
		}
	}
}
