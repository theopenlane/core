package keymaker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/samber/lo"

	"github.com/theopenlane/common/integrations/types"
	"github.com/theopenlane/core/pkg/integrations"
)

// defaultSessionTTL is the duration that OAuth sessions remain valid if no custom TTL is configured
const defaultSessionTTL = 15 * time.Minute

// ProviderResolver exposes provider lookups. registry.Registry satisfies this interface
type ProviderResolver interface {
	Provider(provider types.ProviderType) (types.Provider, bool)
}

// CredentialWriter persists credential payloads produced during activation
type CredentialWriter interface {
	SaveCredential(ctx context.Context, orgID string, payload types.CredentialPayload) (types.CredentialPayload, error)
}

// ServiceOptions configure optional service behaviors
type ServiceOptions struct {
	// SessionTTL controls how long OAuth sessions remain valid
	SessionTTL time.Duration
	// Now overrides the time source; primarily used for tests
	Now func() time.Time
}

// Service orchestrates activation flows by brokering providers, sessions, and keystore writes
type Service struct {
	// providers resolves provider instances by type
	providers ProviderResolver
	// keystore persists credential payloads after activation
	keystore CredentialWriter
	// sessions stores temporary OAuth state until callback completion
	sessions SessionStore

	// sessionTTL controls the lifetime of pending OAuth sessions
	sessionTTL time.Duration
	// now returns the current time, overridable for testing
	now func() time.Time
}

// NewService constructs a Service from the supplied dependencies
func NewService(providers ProviderResolver, keystore CredentialWriter, sessions SessionStore, opts ServiceOptions) (*Service, error) {
	if providers == nil {
		return nil, integrations.ErrProviderRegistryUninitialized
	}

	if keystore == nil {
		return nil, integrations.ErrKeystoreRequired
	}

	if sessions == nil {
		return nil, integrations.ErrSessionStoreRequired
	}

	ttl := opts.SessionTTL
	if ttl <= 0 {
		ttl = defaultSessionTTL
	}

	nowFn := opts.Now
	if nowFn == nil {
		nowFn = time.Now
	}

	return &Service{
		providers:  providers,
		keystore:   keystore,
		sessions:   sessions,
		sessionTTL: ttl,
		now:        nowFn,
	}, nil
}

// BeginRequest carries the information required to start an OAuth/OIDC activation flow
type BeginRequest struct {
	// OrgID identifies the organization initiating the flow
	OrgID string
	// IntegrationID identifies the integration record being activated
	IntegrationID string
	// Provider specifies which provider to use for authorization
	Provider types.ProviderType
	// RedirectURI overrides the default callback URL if specified
	RedirectURI string
	// Scopes requests specific authorization scopes from the provider
	Scopes []string
	// Metadata carries additional provider-specific configuration
	Metadata map[string]any
	// LabelOverrides customizes UI labels presented during authorization
	LabelOverrides map[string]string
	// State optionally supplies a custom CSRF token
	State string
}

// BeginResponse returns the authorization URL/state pair for the caller to redirect the user
type BeginResponse struct {
	// Provider identifies which provider is handling the authorization
	Provider types.ProviderType
	// State contains the CSRF token that must be validated during callback
	State string
	// AuthURL is the provider authorization URL where the user should be redirected
	AuthURL string
}

// CompleteRequest carries the state/code pair received from the provider callback
type CompleteRequest struct {
	// State is the CSRF token returned by the provider that identifies the session
	State string
	// Code is the authorization code exchanged for credentials
	Code string
}

// CompleteResult reports the persisted credential and related identifiers
type CompleteResult struct {
	// Provider identifies which provider issued the credential
	Provider types.ProviderType
	// OrgID identifies the organization that owns the credential
	OrgID string
	// IntegrationID identifies the integration record containing the credential
	IntegrationID string
	// Credential contains the persisted credential payload
	Credential types.CredentialPayload
}

// BeginAuthorization starts an OAuth/OIDC transaction with the requested provider
func (s *Service) BeginAuthorization(ctx context.Context, req BeginRequest) (BeginResponse, error) {
	if strings.TrimSpace(req.OrgID) == "" {
		return BeginResponse{}, integrations.ErrOrgIDRequired
	}

	if strings.TrimSpace(req.IntegrationID) == "" {
		return BeginResponse{}, integrations.ErrIntegrationIDRequired
	}

	if req.Provider == types.ProviderUnknown {
		return BeginResponse{}, types.ErrProviderTypeRequired
	}

	provider, ok := s.providers.Provider(req.Provider)
	if !ok {
		return BeginResponse{}, integrations.ErrProviderNotFound
	}

	authCtx := types.AuthContext{
		OrgID:          req.OrgID,
		IntegrationID:  req.IntegrationID,
		RedirectURI:    req.RedirectURI,
		State:          strings.TrimSpace(req.State),
		Scopes:         append([]string(nil), req.Scopes...),
		Metadata:       lo.Assign(map[string]any{}, req.Metadata),
		LabelOverrides: lo.Assign(map[string]string{}, req.LabelOverrides),
	}

	session, err := provider.BeginAuth(ctx, authCtx)
	if err != nil {
		return BeginResponse{}, fmt.Errorf("keymaker: begin auth: %w", err)
	}

	state := strings.TrimSpace(session.State())
	if state == "" {
		return BeginResponse{}, integrations.ErrStateRequired
	}

	activation := ActivationSession{
		State:          state,
		Provider:       req.Provider,
		OrgID:          req.OrgID,
		IntegrationID:  req.IntegrationID,
		Scopes:         append([]string(nil), req.Scopes...),
		Metadata:       lo.Assign(map[string]any{}, req.Metadata),
		LabelOverrides: lo.Assign(map[string]string{}, req.LabelOverrides),
		AuthSession:    session,
		CreatedAt:      s.now(),
	}

	activation.ExpiresAt = activation.CreatedAt.Add(s.sessionTTL)

	if err := s.sessions.Save(activation); err != nil {
		return BeginResponse{}, fmt.Errorf("keymaker: save auth session: %w", err)
	}

	return BeginResponse{
		Provider: req.Provider,
		State:    state,
		AuthURL:  session.AuthURL(),
	}, nil
}

// CompleteAuthorization finalizes an OAuth/OIDC transaction and persists the resulting credential
func (s *Service) CompleteAuthorization(ctx context.Context, req CompleteRequest) (CompleteResult, error) {
	if strings.TrimSpace(req.State) == "" {
		return CompleteResult{}, integrations.ErrStateRequired
	}
	if strings.TrimSpace(req.Code) == "" {
		return CompleteResult{}, integrations.ErrAuthorizationCodeRequired
	}

	activation, err := s.sessions.Take(req.State)
	if err != nil {
		return CompleteResult{}, err
	}

	if s.now().After(activation.ExpiresAt) {
		return CompleteResult{}, integrations.ErrAuthorizationStateExpired
	}

	if activation.AuthSession == nil {
		return CompleteResult{}, integrations.ErrAuthSessionInvalid
	}

	payload, err := activation.AuthSession.Finish(ctx, req.Code)
	if err != nil {
		return CompleteResult{}, fmt.Errorf("keymaker: finish auth: %w", err)
	}

	if payload.Provider == types.ProviderUnknown {
		payload.Provider = activation.Provider
	}

	saved, err := s.keystore.SaveCredential(ctx, activation.OrgID, payload)
	if err != nil {
		return CompleteResult{}, fmt.Errorf("keymaker: save credential: %w", err)
	}

	return CompleteResult{
		Provider:      activation.Provider,
		OrgID:         activation.OrgID,
		IntegrationID: activation.IntegrationID,
		Credential:    saved,
	}, nil
}
