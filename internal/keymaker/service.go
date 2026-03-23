package keymaker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// DefinitionLookupFunc resolves definitions for auth flow dispatch
type DefinitionLookupFunc func(id string) (types.Definition, bool)

// AuthCompleteHookFunc is the callback invoked after a definition auth flow completes successfully
type AuthCompleteHookFunc func(ctx context.Context, installationID string, credentialRef types.CredentialSlotID, definition types.Definition, result types.AuthCompleteResult) error

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
	// onAuthComplete is the callback invoked after activation completes
	onAuthComplete AuthCompleteHookFunc
	// installationLookup resolves and validates installation records referenced by auth flows
	installationLookup InstallationLookupFunc
	// authStates stores temporary authorization state until callback completion
	authStates AuthStateStore
}

// NewService constructs a Service from the supplied dependencies
func NewService(definitionLookup DefinitionLookupFunc, onAuthComplete AuthCompleteHookFunc, installationLookup InstallationLookupFunc, authStates AuthStateStore) *Service {
	return &Service{
		definitionLookup:   definitionLookup,
		onAuthComplete:     onAuthComplete,
		installationLookup: installationLookup,
		authStates:         authStates,
	}
}

// BeginRequest carries the information required to start a definition auth flow
type BeginRequest struct {
	// DefinitionID identifies which definition to use for authorization
	DefinitionID string
	// InstallationID identifies the installation record being activated
	InstallationID string
	// CredentialRef identifies which credential-schema-selected connection mode should be activated
	CredentialRef types.CredentialSlotID
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
	// Callback carries the generic callback payload captured from the provider redirect
	Callback types.AuthCallbackInput
}

// CompleteResult reports the persisted credential and related identifiers
type CompleteResult struct {
	// DefinitionID identifies which definition issued the credential
	DefinitionID string
	// InstallationID identifies the installation record containing the credential
	InstallationID string
	// CredentialRef identifies which credential slot received the persisted credential
	CredentialRef types.CredentialSlotID
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

	connection, err := def.ConnectionRegistration(req.CredentialRef)
	if err != nil {
		return BeginResponse{}, ErrConnectionNotFound
	}

	if connection.Auth == nil || connection.Auth.Start == nil {
		return BeginResponse{}, ErrDefinitionAuthRequired
	}

	if err := s.validateInstallation(ctx, req.InstallationID, req.DefinitionID); err != nil {
		return BeginResponse{}, err
	}

	result, err := connection.Auth.Start(ctx, jsonx.CloneRawMessage(req.Input))
	if err != nil {
		return BeginResponse{}, fmt.Errorf("keymaker: begin definition auth: %w", err)
	}

	state, err := auth.GenerateOAuthState(0)
	if err != nil {
		return BeginResponse{}, fmt.Errorf("keymaker: generate session state: %w", err)
	}

	authState := AuthState{
		State:          state,
		DefinitionID:   req.DefinitionID,
		InstallationID: req.InstallationID,
		CredentialRef:  req.CredentialRef,
		CallbackState:  jsonx.CloneRawMessage(result.State),
	}

	if err := s.authStates.Save(authState); err != nil {
		return BeginResponse{}, fmt.Errorf("keymaker: save definition auth session: %w", err)
	}

	return BeginResponse{
		DefinitionID: req.DefinitionID,
		State:        state,
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

	def, ok := s.definitionLookup(authState.DefinitionID)
	if !ok {
		return CompleteResult{}, ErrDefinitionNotFound
	}

	connection, err := def.ConnectionRegistration(authState.CredentialRef)
	if err != nil {
		return CompleteResult{}, ErrConnectionNotFound
	}

	if connection.Auth == nil || connection.Auth.Complete == nil {
		return CompleteResult{}, ErrDefinitionAuthRequired
	}

	if err := s.validateInstallation(ctx, authState.InstallationID, authState.DefinitionID); err != nil {
		return CompleteResult{}, err
	}

	completeResult, err := connection.Auth.Complete(ctx, jsonx.CloneRawMessage(authState.CallbackState), req.Callback)
	if err != nil {
		return CompleteResult{}, fmt.Errorf("keymaker: complete definition auth: %w", err)
	}

	if err := s.onAuthComplete(ctx, authState.InstallationID, authState.CredentialRef, def, completeResult); err != nil {
		return CompleteResult{}, fmt.Errorf("keymaker: auth complete hook: %w", err)
	}

	return CompleteResult{
		DefinitionID:   authState.DefinitionID,
		InstallationID: authState.InstallationID,
		CredentialRef:  connection.Auth.CredentialRef,
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
