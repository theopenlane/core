package keymaker

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/theopenlane/core/internal/integrations"
	"github.com/theopenlane/core/internal/integrations/types"
)

// ActivationSession captures the temporary state required to complete an OAuth flow
type ActivationSession struct {
	State          string
	Provider       types.ProviderType
	OrgID          string
	IntegrationID  string
	Scopes         []string
	Metadata       map[string]any
	LabelOverrides map[string]string
	CreatedAt      time.Time
	ExpiresAt      time.Time
	AuthSession    types.AuthSession
}

// SessionStore persists activation sessions until the provider callback is completed
type SessionStore interface {
	Save(ctx context.Context, session ActivationSession) error
	Take(ctx context.Context, state string) (ActivationSession, error)
}

// MemorySessionStore stores activation sessions in memory and is safe for concurrent use
type MemorySessionStore struct {
	mu       sync.Mutex
	sessions map[string]ActivationSession
}

// NewMemorySessionStore returns an in-memory session store
func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{
		sessions: map[string]ActivationSession{},
	}
}

// Save records the provided activation session
func (m *MemorySessionStore) Save(_ context.Context, session ActivationSession) error {
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
func (m *MemorySessionStore) Take(_ context.Context, state string) (ActivationSession, error) {
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
