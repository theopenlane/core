package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/samber/lo"

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
func (r *Runtime) LoadCredential(ctx context.Context, installation *ent.Integration, credentialRef types.CredentialSlotID) (types.CredentialSet, bool, error) {
	return r.Keystore().LoadCredential(ctx, installation, credentialRef)
}

// LoadCredentials resolves the requested credential slots for one installation
func (r *Runtime) LoadCredentials(ctx context.Context, installation *ent.Integration, credentialRefs []types.CredentialSlotID) (types.CredentialBindings, error) {
	return r.Keystore().LoadCredentials(ctx, installation, credentialRefs)
}

// DeleteCredential removes credentials for one installation identifier and evicts cached clients
func (r *Runtime) DeleteCredential(ctx context.Context, installationID string) error {
	return r.Keystore().DeleteCredential(ctx, installationID)
}

// cleanupInstallation removes credentials and the installation record for one installation
func (r *Runtime) cleanupInstallation(ctx context.Context, installationID string) error {
	if err := r.Keystore().DeleteCredential(ctx, installationID); err != nil {
		return err
	}

	return r.DB().Integration.DeleteOneID(installationID).Exec(ctx)
}

// Disconnect executes the teardown flow for one installation
func (r *Runtime) Disconnect(ctx context.Context, installation *ent.Integration) (types.DisconnectResult, error) {
	def, err := r.resolveDefinitionForInstallation(installation)
	if err != nil {
		return types.DisconnectResult{}, err
	}

	connection, err := r.resolvePersistedConnection(def, installation)
	if err != nil && !errors.Is(err, ErrConnectionRequired) {
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

	if err := r.cleanupInstallation(ctx, installation.ID); err != nil {
		return types.DisconnectResult{}, err
	}

	return result, nil
}

// Reconcile reconciles installation user input and/or one credential update
func (r *Runtime) Reconcile(ctx context.Context, installation *ent.Integration, userInput json.RawMessage, credentialRef types.CredentialSlotID, credential *types.CredentialSet, installationInput json.RawMessage) error {
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
	if def.UserInput != nil {
		if err := validatePayload(def.UserInput.Schema, userInput, ErrUserInputInvalid); err != nil {
			return err
		}
	}

	installation.Config.ClientConfig = jsonx.CloneRawMessage(userInput)

	update := r.DB().Integration.UpdateOneID(installation.ID).SetConfig(installation.Config)

	decoded := jsonx.DecodeAnyOrNil(userInput)
	if m, ok := decoded.(map[string]any); ok {
		if name, ok := m["name"].(string); ok && name != "" {
			update.SetName(name)
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

	if state.CredentialRef == (types.CredentialSlotID{}) {
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

	if connection.Installation == nil {
		return nil
	}

	credential, _ := bindings.Resolve(connection.CredentialRef)

	metadata, ok, err := connection.Installation.Resolve(systemCtx, types.InstallationRequest{
		Installation: installation,
		Connection:   connection,
		Credential:   credential,
		Credentials:  bindings,
		Config:       installation.Config,
	})
	if err != nil {
		return err
	}

	if !ok {
		return r.saveInstallationMetadata(systemCtx, installation, types.IntegrationInstallationMetadata{})
	}

	return r.saveInstallationMetadata(systemCtx, installation, metadata)
}

// reconcileCredential validates, health-checks, and persists one credential for an installation
func (r *Runtime) reconcileCredential(ctx context.Context, installation *ent.Integration, def types.Definition, credentialRef types.CredentialSlotID, credential types.CredentialSet, installationInput json.RawMessage) error {
	registration, err := def.CredentialRegistration(credentialRef)
	if err != nil {
		return err
	}

	connection, err := r.resolveConnectionForCredential(def, installation, credentialRef)
	if err != nil {
		return err
	}

	if err := validatePayload(registration.Schema, credential.Data, ErrCredentialInvalid); err != nil {
		return err
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
			logx.FromContext(ctx).Error().Err(validationErr).Msg("validation failed during reconcile")

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

	if connection.Installation != nil {
		metadata, ok, err := connection.Installation.Resolve(systemCtx, types.InstallationRequest{
			Installation: installation,
			Connection:   connection,
			Credential:   credential,
			Credentials:  bindings,
			Config:       installation.Config,
			Input:        installationInput,
		})
		if err != nil {
			return err
		}

		if !ok {
			metadata = types.IntegrationInstallationMetadata{}
		}

		if err := r.saveInstallationMetadata(systemCtx, installation, metadata); err != nil {
			return err
		}
	}

	if err := r.DB().Integration.UpdateOneID(installation.ID).
		SetStatus(enums.IntegrationStatusConnected).
		Exec(systemCtx); err != nil {
		return err
	}

	installation.Status = enums.IntegrationStatusConnected

	return r.reconcileInstallationWebhooks(systemCtx, installation, "")
}

// resolveConnectionFromState resolves the connection persisted in provider state for an installation
func (r *Runtime) resolveConnectionFromState(def types.Definition, installation *ent.Integration) (types.ConnectionRegistration, bool, error) {
	state, err := def.ProviderState(installation.ProviderState)
	if err != nil {
		return types.ConnectionRegistration{}, false, err
	}

	if state.CredentialRef == (types.CredentialSlotID{}) {
		return types.ConnectionRegistration{}, false, nil
	}

	connection, err := def.ConnectionRegistration(state.CredentialRef)
	if err != nil {
		return types.ConnectionRegistration{}, false, ErrConnectionNotFound
	}

	return connection, true, nil
}

// resolvePersistedConnection resolves the persisted connection for an installation
func (r *Runtime) resolvePersistedConnection(def types.Definition, installation *ent.Integration) (types.ConnectionRegistration, error) {
	connection, found, err := r.resolveConnectionFromState(def, installation)
	if err != nil {
		return types.ConnectionRegistration{}, err
	}

	if found {
		return connection, nil
	}

	if len(def.Connections) == 1 {
		return def.Connections[0], nil
	}

	return types.ConnectionRegistration{}, ErrConnectionRequired
}

// resolveConnectionForCredential resolves the connection for a given credential reference
func (r *Runtime) resolveConnectionForCredential(def types.Definition, installation *ent.Integration, credentialRef types.CredentialSlotID) (types.ConnectionRegistration, error) {
	connection, found, err := r.resolveConnectionFromState(def, installation)
	if err != nil {
		return types.ConnectionRegistration{}, err
	}

	if found {
		if !lo.Contains(connection.CredentialRefs, credentialRef) {
			return types.ConnectionRegistration{}, ErrCredentialNotDeclared
		}

		return connection, nil
	}

	if credentialRef == (types.CredentialSlotID{}) {
		return types.ConnectionRegistration{}, ErrConnectionRequired
	}

	connection, err = def.ConnectionRegistration(credentialRef)
	if err != nil {
		return types.ConnectionRegistration{}, ErrConnectionNotFound
	}

	return connection, nil
}

// candidateCredentials returns credential bindings with a candidate credential override
func (r *Runtime) candidateCredentials(ctx context.Context, installation *ent.Integration, connection types.ConnectionRegistration, credentialRef types.CredentialSlotID, credential types.CredentialSet) (types.CredentialBindings, error) {
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

// connectionCredentials loads all credentials for a connection
func (r *Runtime) connectionCredentials(ctx context.Context, installation *ent.Integration, connection types.ConnectionRegistration) (types.CredentialBindings, error) {
	return r.LoadCredentials(ctx, installation, connection.CredentialRefs)
}

// persistConnectionState updates the provider state for an installation with a new credential reference
func (r *Runtime) persistConnectionState(ctx context.Context, installation *ent.Integration, def types.Definition, credentialRef types.CredentialSlotID) error {
	next, err := def.WithProviderState(installation.ProviderState, types.DefinitionProviderState{
		CredentialRef: credentialRef,
	})
	if err != nil {
		return err
	}

	return r.DB().Integration.UpdateOneID(installation.ID).SetProviderState(next).Exec(ctx)
}
