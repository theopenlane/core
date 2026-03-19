package keymaker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// stateTokenEntropyBytes is the number of random bytes used for CSRF state token generation
const stateTokenEntropyBytes = 16

// defaultSessionTTL is the duration that auth sessions remain valid if no custom TTL is configured
const defaultSessionTTL = 15 * time.Minute

// DefinitionLookupFunc resolves definitions for auth flow dispatch.
type DefinitionLookupFunc func(id string) (types.Definition, bool)

// PersistAuthResultFunc persists auth completion payloads produced during definition auth activation.
type PersistAuthResultFunc func(ctx context.Context, installationID string, definition types.Definition, result types.AuthCompleteResult) error

// InstallationRecord captures the installation fields required by auth validation.
type InstallationRecord struct {
	ID           string
	OwnerID      string
	DefinitionID string
}

// InstallationLookupFunc resolves one installation used during auth flow validation.
type InstallationLookupFunc func(ctx context.Context, installationID string) (InstallationRecord, error)

// Service orchestrates auth flows for integration definitions
type Service struct {
	// definitionLookup resolves definition instances by ID
	definitionLookup DefinitionLookupFunc
	// persistAuthResult persists completion output after activation
	persistAuthResult PersistAuthResultFunc
	// installationLookup resolves and validates installation records referenced by auth flows
	installationLookup InstallationLookupFunc
	// authStates stores temporary authorization state until callback completion
	authStates AuthStateStore
	// sessionTTL controls the lifetime of pending auth sessions
	sessionTTL time.Duration
	// now returns the current time, overridable for testing
	now func() time.Time
}

// NewService constructs a Service from the supplied dependencies
func NewService(definitionLookup DefinitionLookupFunc, persistAuthResult PersistAuthResultFunc, installationLookup InstallationLookupFunc, authStates AuthStateStore, sessionTTL time.Duration) *Service {
	if sessionTTL <= 0 {
		sessionTTL = defaultSessionTTL
	}

	return &Service{
		definitionLookup:   definitionLookup,
		persistAuthResult:  persistAuthResult,
		installationLookup: installationLookup,
		authStates:         authStates,
		sessionTTL:         sessionTTL,
		now:                time.Now,
	}
}

// BeginRequest carries the information required to start a definition auth flow
type BeginRequest struct {
	// DefinitionID identifies which definition to use for authorization
	DefinitionID string
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
	DefinitionID string
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
	DefinitionID string
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

	def, ok := s.definitionLookup(req.DefinitionID)
	if !ok {
		return BeginResponse{}, ErrDefinitionNotFound
	}

	if def.Auth == nil || (def.Auth.Start == nil && def.Auth.OAuth == nil) {
		return BeginResponse{}, ErrDefinitionAuthRequired
	}

	if err := s.validateInstallation(ctx, req.InstallationID, req.DefinitionID); err != nil {
		return BeginResponse{}, err
	}

	result, err := s.beginDefinitionAuth(ctx, def, jsonx.CloneRawMessage(req.Input))
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

// CompleteAuth finalizes a definition auth transaction and persists the resulting auth result
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

	def, ok := s.definitionLookup(authState.DefinitionID)
	if !ok {
		return CompleteResult{}, ErrDefinitionNotFound
	}

	if def.Auth == nil || (def.Auth.Complete == nil && def.Auth.OAuth == nil) {
		return CompleteResult{}, ErrDefinitionAuthRequired
	}

	if err := s.validateInstallation(ctx, authState.InstallationID, authState.DefinitionID); err != nil {
		return CompleteResult{}, err
	}

	completeResult, err := s.completeDefinitionAuth(ctx, def, jsonx.CloneRawMessage(authState.CallbackState), jsonx.CloneRawMessage(req.Input))
	if err != nil {
		return CompleteResult{}, fmt.Errorf("keymaker: complete definition auth: %w", err)
	}

	if err := s.persistAuthResult(ctx, authState.InstallationID, def, completeResult); err != nil {
		return CompleteResult{}, fmt.Errorf("keymaker: persist definition auth result: %w", err)
	}

	return CompleteResult{
		DefinitionID:   authState.DefinitionID,
		InstallationID: authState.InstallationID,
		Credential:     completeResult.Credential,
	}, nil
}

func (s *Service) validateInstallation(ctx context.Context, installationID string, definitionID string) error {
	installation, err := s.installationLookup(ctx, installationID)
	if err != nil {
		return err
	}

	if installation.DefinitionID != definitionID {
		return ErrInstallationDefinitionMismatch
	}

	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil || caller.OrganizationID == "" {
		return nil
	}

	if installation.OwnerID != caller.OrganizationID {
		return ErrInstallationOwnerMismatch
	}

	return nil
}

func (s *Service) beginDefinitionAuth(ctx context.Context, def types.Definition, input json.RawMessage) (types.AuthStartResult, error) {
	if def.Auth.Start != nil {
		return def.Auth.Start(ctx, input)
	}

	return providerkit.StartOAuthFlow(ctx, providerkit.OAuthFlowConfig{
		ClientID:     def.Auth.OAuth.ClientID,
		ClientSecret: def.Auth.ClientSecret,
		AuthURL:      def.Auth.OAuth.AuthURL,
		TokenURL:     def.Auth.OAuth.TokenURL,
		DiscoveryURL: def.Auth.DiscoveryURL,
		RedirectURL:  def.Auth.OAuth.RedirectURI,
		Scopes:       def.Auth.OAuth.Scopes,
		AuthParams:   def.Auth.OAuth.AuthParams,
		TokenParams:  def.Auth.OAuth.TokenParams,
	})
}

func (s *Service) completeDefinitionAuth(ctx context.Context, def types.Definition, state json.RawMessage, input json.RawMessage) (types.AuthCompleteResult, error) {
	if def.Auth.Complete != nil {
		return def.Auth.Complete(ctx, state, input)
	}

	return providerkit.CompleteOAuthFlow(ctx, providerkit.OAuthFlowConfig{
		ClientID:     def.Auth.OAuth.ClientID,
		ClientSecret: def.Auth.ClientSecret,
		AuthURL:      def.Auth.OAuth.AuthURL,
		TokenURL:     def.Auth.OAuth.TokenURL,
		DiscoveryURL: def.Auth.DiscoveryURL,
		RedirectURL:  def.Auth.OAuth.RedirectURI,
		Scopes:       def.Auth.OAuth.Scopes,
		AuthParams:   def.Auth.OAuth.AuthParams,
		TokenParams:  def.Auth.OAuth.TokenParams,
	}, state, input)
}
