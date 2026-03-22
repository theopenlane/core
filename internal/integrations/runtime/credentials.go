package runtime

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/keymaker"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// BeginAuth starts one definition auth flow through the runtime-managed keymaker service
func (r *Runtime) BeginAuth(ctx context.Context, req keymaker.BeginRequest) (keymaker.BeginResponse, error) {
	return r.Keymaker().BeginAuth(ctx, req)
}

// CompleteAuth completes one definition auth flow through the runtime-managed keymaker service
func (r *Runtime) CompleteAuth(ctx context.Context, req keymaker.CompleteRequest) (keymaker.CompleteResult, error) {
	return r.Keymaker().CompleteAuth(ctx, req)
}

// LoadCredential resolves one persisted credential slot for one installation
func (r *Runtime) LoadCredential(ctx context.Context, installation *ent.Integration, credentialRef types.CredentialRef) (types.CredentialSet, bool, error) {
	return r.Keystore().LoadCredential(ctx, installation, credentialRef)
}

// LoadCredentials resolves the requested credential slots for one installation
func (r *Runtime) LoadCredentials(ctx context.Context, installation *ent.Integration, credentialRefs []types.CredentialRef) (types.CredentialBindings, error) {
	return r.Keystore().LoadCredentials(ctx, installation, credentialRefs)
}

// DeleteCredential removes credentials for one installation identifier and evicts cached clients
func (r *Runtime) DeleteCredential(ctx context.Context, installationID string) error {
	return r.Keystore().DeleteCredential(ctx, installationID)
}

// Disconnect executes the teardown flow for one installation
func (r *Runtime) Disconnect(ctx context.Context, installation *ent.Integration) (types.DisconnectResult, error) {
	def, err := r.resolveDefinitionForInstallation(installation)
	if err != nil {
		return types.DisconnectResult{}, err
	}

	connection, err := r.resolvePersistedConnection(def, installation)
	if err != nil && err != ErrConnectionRequired {
		return types.DisconnectResult{}, err
	}

	var result types.DisconnectResult

	if err == nil && connection.Disconnect != nil && connection.Disconnect.Disconnect != nil {
		credentials, loadErr := r.connectionCredentials(ctx, installation, connection)
		if loadErr != nil {
			return types.DisconnectResult{}, loadErr
		}

		var credential *types.CredentialSet
		if resolved, ok := credentials.Resolve(connection.Disconnect.CredentialRef); ok {
			credential = &resolved
		}

		result, err = connection.Disconnect.Disconnect(ctx, types.DisconnectRequest{
			Installation: installation,
			Connection:   connection,
			Credential:   credential,
			Credentials:  credentials,
			Config:       installation.Config,
		})
		if err != nil {
			return types.DisconnectResult{}, err
		}

		if result.SkipLocalCleanup {
			return result, nil
		}
	}

	if err := r.Keystore().DeleteCredential(ctx, installation.ID); err != nil {
		return types.DisconnectResult{}, err
	}

	if err := r.DB().Integration.DeleteOneID(installation.ID).Exec(ctx); err != nil {
		return types.DisconnectResult{}, err
	}

	return result, nil
}

// Reconcile reconciles installation user input and/or one credential update
func (r *Runtime) Reconcile(
	ctx context.Context,
	installation *ent.Integration,
	userInput json.RawMessage,
	credentialRef types.CredentialRef,
	credential *types.CredentialSet,
	installationInput json.RawMessage,
) error {
	def, err := r.resolveDefinitionForInstallation(installation)
	if err != nil {
		return err
	}

	if !jsonx.IsEmptyRawMessage(userInput) {
		if err := r.reconcileUserInput(ctx, installation, def, userInput); err != nil {
			return err
		}
	}

	if credential != nil {
		if err := r.reconcileCredential(ctx, installation, def, credentialRef, *credential, installationInput); err != nil {
			return err
		}
	}

	return nil
}

// reconcileUserInput validates and persists user input for one installation
func (r *Runtime) reconcileUserInput(ctx context.Context, installation *ent.Integration, def types.Definition, userInput json.RawMessage) error {
	if def.UserInput != nil && len(def.UserInput.Schema) > 0 {
		result, err := jsonx.ValidateSchema(def.UserInput.Schema, userInput)
		if err != nil {
			return fmt.Errorf("reconcile: validate user input schema: %w", err)
		}

		if !result.Valid() {
			return ErrUserInputInvalid
		}
	}

	installation.Config.ClientConfig = jsonx.CloneRawMessage(userInput)

	update := r.DB().Integration.UpdateOneID(installation.ID).SetConfig(installation.Config)

	decoded := jsonx.DecodeAnyOrNil(userInput)
	if m, ok := decoded.(map[string]any); ok {
		if name, ok := m["name"].(string); ok && name != "" {
			update.SetName(name)
			installation.Name = name
		}
	}

	if err := update.Exec(ctx); err != nil {
		return err
	}

	systemCtx := privacy.DecisionContext(ctx, privacy.Allow)

	state, err := def.ProviderState(installation.ProviderState)
	if err != nil {
		return err
	}

	if state.CredentialRef == (types.CredentialRef{}) {
		return nil
	}

	connection, err := def.ConnectionRegistration(state.CredentialRef)
	if err != nil {
		return err
	}

	bindings, err := r.connectionCredentials(systemCtx, installation, connection)
	if err != nil {
		return err
	}

	credential, _ := bindings.Resolve(connection.CredentialRef)

	return r.reconcileConnectionInstallationMetadata(systemCtx, installation, connection, credential, bindings, nil)
}

// reconcileCredential validates, health-checks, and persists one credential for an installation
func (r *Runtime) reconcileCredential(
	ctx context.Context,
	installation *ent.Integration,
	def types.Definition,
	credentialRef types.CredentialRef,
	credential types.CredentialSet,
	installationInput json.RawMessage,
) error {
	log := logx.FromContext(ctx)

	registration, err := def.CredentialRegistration(credentialRef)
	if err != nil {
		return err
	}

	connection, err := r.resolveConnectionForCredential(def, installation, credentialRef)
	if err != nil {
		return err
	}

	if len(registration.Schema) > 0 {
		result, err := jsonx.ValidateSchema(registration.Schema, credential.Data)
		if err != nil {
			return fmt.Errorf("reconcile: validate credential schema: %w", err)
		}

		if !result.Valid() {
			return ErrCredentialInvalid
		}
	}

	bindings, err := r.candidateCredentials(ctx, installation, connection, credentialRef, credential)
	if err != nil {
		return err
	}

	if connection.ValidationOperation != "" {
		validationOp, err := r.Registry().Operation(def.ID, connection.ValidationOperation)
		if err != nil {
			return fmt.Errorf("reconcile: resolve validation operation: %w", err)
		}

		_, validationErr := r.ExecuteOperation(ctx, installation, validationOp, bindings, nil)
		if validationErr != nil {
			log.Warn().Err(validationErr).Str("installation_id", installation.ID).Str("credential_ref", connection.CredentialRef.String()).Msg("validation failed during reconcile")
			return fmt.Errorf("reconcile: validation failed: %w", validationErr)
		}
	}

	systemCtx := privacy.DecisionContext(ctx, privacy.Allow)

	if err := r.Keystore().SaveCredential(systemCtx, installation, registration.Ref, credential); err != nil {
		return err
	}

	if err := r.persistConnectionState(systemCtx, installation, def, connection.CredentialRef); err != nil {
		return err
	}

	if err := r.reconcileConnectionInstallationMetadata(systemCtx, installation, connection, credential, bindings, installationInput); err != nil {
		return err
	}

	if err := r.DB().Integration.UpdateOneID(installation.ID).
		SetStatus(enums.IntegrationStatusConnected).
		Exec(systemCtx); err != nil {
		return err
	}

	installation.Status = enums.IntegrationStatusConnected

	return r.reconcileInstallationWebhooks(systemCtx, installation, "")
}

func (r *Runtime) reconcileConnectionInstallationMetadata(
	ctx context.Context,
	installation *ent.Integration,
	connection types.ConnectionRegistration,
	credential types.CredentialSet,
	bindings types.CredentialBindings,
	input json.RawMessage,
) error {
	if connection.Installation == nil {
		return nil
	}

	metadata, ok, err := connection.Installation.Resolve(ctx, types.InstallationRequest{
		Installation: installation,
		Connection:   connection,
		Credential:   credential,
		Credentials:  bindings,
		Config:       installation.Config,
		Input:        input,
	})
	if err != nil {
		return err
	}

	if !ok {
		return saveInstallationMetadata(ctx, installation, types.IntegrationInstallationMetadata{})
	}

	return saveInstallationMetadata(ctx, installation, metadata)
}

func (r *Runtime) resolvePersistedConnection(def types.Definition, installation *ent.Integration) (types.ConnectionRegistration, error) {
	state, err := def.ProviderState(installation.ProviderState)
	if err != nil {
		return types.ConnectionRegistration{}, err
	}

	if state.CredentialRef != (types.CredentialRef{}) {
		connection, err := def.ConnectionRegistration(state.CredentialRef)
		if err != nil {
			return types.ConnectionRegistration{}, ErrConnectionNotFound
		}

		return connection, nil
	}

	if len(def.Connections) == 1 {
		return def.Connections[0], nil
	}

	return types.ConnectionRegistration{}, ErrConnectionRequired
}

func (r *Runtime) resolveConnectionForCredential(def types.Definition, installation *ent.Integration, credentialRef types.CredentialRef) (types.ConnectionRegistration, error) {
	state, err := def.ProviderState(installation.ProviderState)
	if err != nil {
		return types.ConnectionRegistration{}, err
	}

	if state.CredentialRef != (types.CredentialRef{}) {
		connection, err := def.ConnectionRegistration(state.CredentialRef)
		if err != nil {
			return types.ConnectionRegistration{}, ErrConnectionNotFound
		}

		credentialDeclared := false
		for _, ref := range connection.CredentialRefs {
			if ref.String() == credentialRef.String() {
				credentialDeclared = true
				break
			}
		}
		if !credentialDeclared {
			return types.ConnectionRegistration{}, ErrConnectionNotFound
		}

		return connection, nil
	}

	if credentialRef == (types.CredentialRef{}) {
		return types.ConnectionRegistration{}, ErrConnectionRequired
	}

	connection, err := def.ConnectionRegistration(credentialRef)
	if err != nil {
		return types.ConnectionRegistration{}, ErrConnectionNotFound
	}

	return connection, nil
}

func (r *Runtime) candidateCredentials(ctx context.Context, installation *ent.Integration, connection types.ConnectionRegistration, credentialRef types.CredentialRef, credential types.CredentialSet) (types.CredentialBindings, error) {
	current, err := r.connectionCredentials(ctx, installation, connection)
	if err != nil {
		return nil, err
	}

	override := types.CredentialBindings{{
		Ref:        credentialRef,
		Credential: credential,
	}}

	return mergeCredentials(current, override), nil
}

func (r *Runtime) connectionCredentials(ctx context.Context, installation *ent.Integration, connection types.ConnectionRegistration) (types.CredentialBindings, error) {
	if len(connection.CredentialRefs) == 0 {
		return nil, nil
	}

	return r.LoadCredentials(ctx, installation, connection.CredentialRefs)
}

func (r *Runtime) persistConnectionState(ctx context.Context, installation *ent.Integration, def types.Definition, credentialRef types.CredentialRef) error {
	next, err := def.WithProviderState(installation.ProviderState, types.DefinitionProviderState{
		CredentialRef: credentialRef,
	})
	if err != nil {
		return err
	}

	if err := r.DB().Integration.UpdateOneID(installation.ID).SetProviderState(next).Exec(ctx); err != nil {
		return err
	}

	installation.ProviderState = next

	return nil
}
