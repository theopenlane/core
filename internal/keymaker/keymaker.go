// Package keymaker manages credential minting and client pooling for integrations.
package keymaker

import (
	"context"
	"time"
)

// Provider issues integration sessions that include hydrated credentials
// and typed clients. Implementations hide persistence, hush access, and
// pooling mechanics behind this interface.
type Provider interface {
	// Session mint a scoped session for the given org/provider pair. The
	// returned session carries fully materialized credentials plus any pooled
	// client references required by downstream callers.
	Session(ctx context.Context, req SessionRequest) (Session, error)

	// Release returns the session to the underlying pool. Callers should invoke
	// this once they are done with the client to avoid leaking connections.
	Release(ctx context.Context, session Session) error
}

// Session bundles provider credentials with an instantiated client handle.
// The client is stored as an opaque interface because specific provider
// packages expose typed accessors (e.g., Slack, GitHub).
type Session struct {
	OrgID         string
	Provider      string
	Credentials   Credentials
	Client        any
	IssuedAt      time.Time
	RefreshAfter  time.Time
	ExpiresAt     time.Time
	RefreshHandle RefreshHandle
	Metadata      map[string]any

	releaseFn func(context.Context) error
}

// Credentials represents the minted token material for a session.
type Credentials struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	Expiry       time.Time
	Scopes       []string
	Raw          map[string]any
}

// RefreshHandle can be attached to a session when the provider supports
// background refresh. It allows the runtime to renew credentials without
// invoking provider-specific code paths.
type RefreshHandle interface {
	Refresh(ctx context.Context) (Credentials, error)
}

// SessionRequest conveys the identifiers and optional overrides needed to
// mint a new session.
type SessionRequest struct {
	OrgID        string
	Integration  string
	Provider     string
	TenantID     string
	Scopes       []string
	ForceRefresh bool
}
