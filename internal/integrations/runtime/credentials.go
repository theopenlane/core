package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/keymaker"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/mapx"
	"github.com/theopenlane/core/pkg/metrics"
)

// BeginAuth starts one definition auth flow through the runtime-managed keymaker service
func (r *Runtime) BeginAuth(ctx context.Context, req keymaker.BeginRequest) (keymaker.BeginResponse, error) {
	return r.keymaker().BeginAuth(ctx, req)
}

// CompleteAuth completes one definition auth flow through the runtime-managed keymaker service
func (r *Runtime) CompleteAuth(ctx context.Context, req keymaker.CompleteRequest) (keymaker.CompleteResult, error) {
	return r.keymaker().CompleteAuth(ctx, req)
}

// LoadCredential resolves one persisted credential slot for one installation
func (r *Runtime) LoadCredential(ctx context.Context, installation *ent.Integration, credentialRef types.CredentialSlotID) (types.CredentialSet, bool, error) {
	return r.keystore().LoadCredential(ctx, installation, credentialRef)
}

// loadCredentials resolves the requested credential slots for one installation
func (r *Runtime) loadCredentials(ctx context.Context, installation *ent.Integration, credentialRefs []types.CredentialSlotID) (types.CredentialBindings, error) {
	return r.keystore().LoadCredentials(ctx, installation, credentialRefs)
}

// deleteCredential removes credentials for one installation identifier and evicts cached clients
func (r *Runtime) deleteCredential(ctx context.Context, integrationID string) error {
	return r.keystore().DeleteCredential(ctx, integrationID)
}

// cleanupInstallation removes credentials and the installation record for one installation
func (r *Runtime) cleanupInstallation(ctx context.Context, integrationID string) error {
	if err := r.deleteCredential(ctx, integrationID); err != nil {
		return err
	}

	return r.DB().Integration.DeleteOneID(integrationID).Exec(ctx)
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
		credentials, loadErr := r.loadCredentials(ctx, installation, connection.CredentialRefs)
		if loadErr != nil {
			return types.DisconnectResult{}, loadErr
		}

		result, err = connection.Disconnect.Disconnect(ctx, types.DisconnectRequest{
			Integration: installation,
			Connection:  connection,
			Credentials: credentials,
			Config:      installation.Config,
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

	metrics.RecordIntegrationDisconnected(installation.DefinitionID)

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

		if primary, ok := m["primaryDirectory"].(bool); ok {
			update.SetPrimaryDirectory(primary)
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

	bindings, err := r.loadCredentials(systemCtx, installation, connection.CredentialRefs)
	if err != nil {
		return err
	}

	if connection.Integration == nil {
		return nil
	}

	metadata, ok, err := connection.Integration.Resolve(systemCtx, types.InstallationRequest{
		Integration: installation,
		Connection:  connection,
		Credentials: bindings,
		Config:      installation.Config,
	})
	if err != nil {
		return err
	}

	if !ok {
		return r.saveInstallationMetadata(systemCtx, installation, types.IntegrationInstallationMetadata{})
	}

	return r.saveInstallationMetadata(systemCtx, installation, metadata)
}

// saveInstallationMetadata persists installation metadata and syncs the normalized
// display identity into the GraphQL-visible metadata map
func (r *Runtime) saveInstallationMetadata(ctx context.Context, installation *ent.Integration, metadata types.IntegrationInstallationMetadata) error {
	displayMeta, _ := jsonx.ToMap(metadata.Display)
	merged := mapx.DeepMergeMapAny(installation.Metadata, mapx.PruneMapZeroAny(displayMeta))

	if err := r.DB().Integration.UpdateOneID(installation.ID).
		SetInstallationMetadata(metadata).
		SetMetadata(merged).
		Exec(ctx); err != nil {
		return err
	}

	installation.InstallationMetadata = metadata
	installation.Metadata = merged

	return nil
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

	bindings, err := r.loadCredentials(ctx, installation, connection.CredentialRefs)
	if err != nil {
		return err
	}

	bindings = bindings.With(credentialRef, credential)

	if connection.ValidationOperation != "" {
		validationOp, err := r.Registry().Operation(def.ID, connection.ValidationOperation)
		if err != nil {
			return fmt.Errorf("resolve validation operation: %w", err)
		}

		_, validationErr := r.ExecuteOperation(ctx, installation, validationOp, bindings, nil)
		if validationErr != nil {
			logx.FromContext(ctx).Error().Err(validationErr).Msg("validation failed during reconcile")

			return fmt.Errorf("validation failed: %w", validationErr)
		}
	}

	systemCtx := privacy.DecisionContext(ctx, privacy.Allow)

	if err := r.keystore().SaveCredential(systemCtx, installation, registration.Ref, credential); err != nil {
		return err
	}

	if err := r.persistConnectionState(systemCtx, installation, def, connection.CredentialRef); err != nil {
		return err
	}

	var metadata types.IntegrationInstallationMetadata

	if connection.Integration != nil {
		resolved, ok, err := connection.Integration.Resolve(systemCtx, types.InstallationRequest{
			Integration: installation,
			Connection:  connection,
			Credentials: bindings,
			Config:      installation.Config,
			Input:       installationInput,
		})
		if err != nil {
			return err
		}

		if ok {
			metadata = resolved
		}
	}

	metadata.Display.CredentialRef = credentialRef.String()

	if err := r.saveInstallationMetadata(systemCtx, installation, metadata); err != nil {
		return err
	}

	wasFirstConnection := installation.Status == enums.IntegrationStatusPending

	if err := r.DB().Integration.UpdateOneID(installation.ID).
		SetStatus(enums.IntegrationStatusConnected).
		Exec(systemCtx); err != nil {
		return err
	}

	installation.Status = enums.IntegrationStatusConnected

	if err := r.reconcileInstallationWebhooks(systemCtx, installation, ""); err != nil {
		return err
	}

	if wasFirstConnection {
		if err := r.reconcileOperations(systemCtx, installation); err != nil {
			return err
		}
	}

	return nil
}

// resolveConnectionFromState resolves the connection persisted in provider state for an installation
func (r *Runtime) resolveConnectionFromState(def types.Definition, installation *ent.Integration) (types.ConnectionRegistration, bool, error) {
	state, err := def.ProviderState(installation.ProviderState)
	if err != nil {
		log.Error().Err(err).Msg("integrations: error getting provider state")
		return types.ConnectionRegistration{}, false, err
	}

	if state.CredentialRef == (types.CredentialSlotID{}) {
		return types.ConnectionRegistration{}, false, nil
	}

	connection, err := def.ConnectionRegistration(state.CredentialRef)
	if err != nil {
		log.Error().Err(err).Msg("integrations: error connecting during registration")

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
		log.Error().Err(err).Msg("integrations: error connecting during registration")
		return types.ConnectionRegistration{}, ErrConnectionNotFound
	}

	return connection, nil
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
