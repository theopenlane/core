package keymaker

import (
	"strings"
	"sync"
	"time"

	"github.com/theopenlane/core/internal/integrations"
	"github.com/theopenlane/core/internal/integrations/types"
)

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
}

// NewMemorySessionStore returns an in-memory session store
func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{
		sessions: map[string]ActivationSession{},
	}
}

// Save records the provided activation session
func (m *MemorySessionStore) Save(session ActivationSession) error {
	if strings.TrimSpace(session.State) == "" {
		return integrations.ErrStateRequired
	}

	if session.AuthSession == nil {
		return integrations.ErrAuthSessionInvalid
	}

	clone := session

	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[session.State] = clone

	return nil
}

// Take retrieves and deletes the session associated with the given state
func (m *MemorySessionStore) Take(state string) (ActivationSession, error) {
	if strings.TrimSpace(state) == "" {
		return ActivationSession{}, integrations.ErrStateRequired
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.sessions[state]
	if !ok {
		return ActivationSession{}, integrations.ErrAuthorizationStateNotFound
	}

	delete(m.sessions, state)

	return session, nil
}
