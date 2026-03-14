package keymaker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// stateTokenEntropyBytes is the number of random bytes used for CSRF state token generation
const stateTokenEntropyBytes = 16

// defaultSessionTTL is the duration that auth sessions remain valid if no custom TTL is configured
const defaultSessionTTL = 15 * time.Minute

// AuthResolver resolves definitions for auth flow dispatch.
// registry.Registry and registry.DefinitionRegistry both satisfy this interface.
type AuthResolver interface {
	Definition(id types.DefinitionID) (types.Definition, bool)
}

// CredentialWriter persists credential payloads produced during definition auth activation.
// keystore.Store satisfies this interface.
type CredentialWriter interface {
	SaveInstallationCredential(ctx context.Context, installationID string, credential types.CredentialSet) error
}

// Options configures optional Service behaviors
type Options struct {
	// SessionTTL controls how long definition auth sessions remain valid
	SessionTTL time.Duration
	// Now overrides the time source; primarily used for tests
	Now func() time.Time
}

// Service orchestrates auth flows for integrationsv2 definitions
type Service struct {
	// definitions resolves definition instances by ID
	definitions AuthResolver
	// writer persists credential payloads after activation
	writer CredentialWriter
	// authStates stores temporary authorization state until callback completion
	authStates AuthStateStore
	// sessionTTL controls the lifetime of pending auth sessions
	sessionTTL time.Duration
	// now returns the current time, overridable for testing
	now func() time.Time
}

// NewService constructs a Service from the supplied dependencies
func NewService(definitions AuthResolver, writer CredentialWriter, authStates AuthStateStore, opts Options) (*Service, error) {
	if definitions == nil {
		return nil, ErrDefinitionResolverRequired
	}

	if writer == nil {
		return nil, ErrCredentialWriterRequired
	}

	if authStates == nil {
		return nil, ErrAuthStateStoreRequired
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
		definitions: definitions,
		writer:      writer,
		authStates:  authStates,
		sessionTTL:  ttl,
		now:         nowFn,
	}, nil
}

// BeginRequest carries the information required to start a definition auth flow
type BeginRequest struct {
	// DefinitionID identifies which definition to use for authorization
	DefinitionID types.DefinitionID
	// InstallationID identifies the installation record being activated
	InstallationID string
	// State optionally supplies a custom CSRF token; one is generated if empty
	State string
	// Input carries optional definition-specific input to the auth start function
	Input json.RawMessage
}

// BeginResponse returns the authorization URL and session state token
type BeginResponse struct {
	// DefinitionID identifies which definition is handling the authorization
	DefinitionID types.DefinitionID
	// State contains the CSRF token that must be presented during callback
	State string
	// AuthURL is the authorization URL where the user should be redirected
	AuthURL string
}

// CompleteRequest carries the state token and callback input from the auth provider
type CompleteRequest struct {
	// State is the CSRF token that identifies the session
	State string
	// Input carries opaque callback data (typically the authorization code and provider state)
	Input json.RawMessage
}

// CompleteResult reports the persisted credential and related identifiers
type CompleteResult struct {
	// DefinitionID identifies which definition issued the credential
	DefinitionID types.DefinitionID
	// InstallationID identifies the installation record containing the credential
	InstallationID string
	// Credential contains the persisted credential payload
	Credential types.CredentialSet
}

// BeginAuth starts an auth transaction for the requested definition
func (s *Service) BeginAuth(ctx context.Context, req BeginRequest) (BeginResponse, error) {
	if req.DefinitionID == "" {
		return BeginResponse{}, ErrDefinitionIDRequired
	}

	if req.InstallationID == "" {
		return BeginResponse{}, ErrInstallationIDRequired
	}

	def, ok := s.definitions.Definition(req.DefinitionID)
	if !ok {
		return BeginResponse{}, ErrDefinitionNotFound
	}

	if def.Auth == nil || def.Auth.Start == nil {
		return BeginResponse{}, ErrDefinitionAuthRequired
	}

	result, err := def.Auth.Start(ctx, jsonx.CloneRawMessage(req.Input))
	if err != nil {
		return BeginResponse{}, fmt.Errorf("keymaker: begin definition auth: %w", err)
	}

	stateToken := req.State
	if stateToken == "" {
		stateToken, err = auth.GenerateOAuthState(stateTokenEntropyBytes)
		if err != nil {
			return BeginResponse{}, fmt.Errorf("keymaker: generate state token: %w", err)
		}
	}

	authState := AuthState{
		State:          stateToken,
		DefinitionID:   req.DefinitionID,
		InstallationID: req.InstallationID,
		CallbackState:  jsonx.CloneRawMessage(result.State),
		CreatedAt:      s.now(),
	}

	authState.ExpiresAt = authState.CreatedAt.Add(s.sessionTTL)

	if err := s.authStates.Save(authState); err != nil {
		return BeginResponse{}, fmt.Errorf("keymaker: save definition auth session: %w", err)
	}

	return BeginResponse{
		DefinitionID: req.DefinitionID,
		State:        stateToken,
		AuthURL:      result.URL,
	}, nil
}

// CompleteAuth finalizes a definition auth transaction and persists the resulting credential
func (s *Service) CompleteAuth(ctx context.Context, req CompleteRequest) (CompleteResult, error) {
	if req.State == "" {
		return CompleteResult{}, ErrAuthStateTokenRequired
	}

	authState, err := s.authStates.Take(req.State)
	if err != nil {
		return CompleteResult{}, err
	}

	if s.now().After(authState.ExpiresAt) {
		return CompleteResult{}, ErrAuthStateExpired
	}

	def, ok := s.definitions.Definition(authState.DefinitionID)
	if !ok {
		return CompleteResult{}, ErrDefinitionNotFound
	}

	if def.Auth == nil || def.Auth.Complete == nil {
		return CompleteResult{}, ErrDefinitionAuthRequired
	}

	completeResult, err := def.Auth.Complete(ctx, jsonx.CloneRawMessage(authState.CallbackState), jsonx.CloneRawMessage(req.Input))
	if err != nil {
		return CompleteResult{}, fmt.Errorf("keymaker: complete definition auth: %w", err)
	}

	if err := s.writer.SaveInstallationCredential(ctx, authState.InstallationID, completeResult.Credential); err != nil {
		return CompleteResult{}, fmt.Errorf("keymaker: save definition credential: %w", err)
	}

	return CompleteResult{
		DefinitionID:   authState.DefinitionID,
		InstallationID: authState.InstallationID,
		Credential:     completeResult.Credential,
	}, nil
}
