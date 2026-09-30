package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	slackdef "github.com/theopenlane/core/v2/internal/integrations/definitions/slack"
	intobvs "github.com/theopenlane/core/v2/internal/integrations/observability"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/internal/keymaker"
	"github.com/theopenlane/core/v2/pkg/jsonx"
	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/core/v2/pkg/mapx"
	"github.com/theopenlane/core/v2/pkg/metrics"
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
	if err := r.ensureCurrentVersion(ctx, installation); err != nil {
		return nil, err
	}

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

// ReapExpiredInstallation soft-deletes an expired never-connected installation and credentials
func (r *Runtime) ReapExpiredInstallation(ctx context.Context, integrationID string) (bool, error) {
	reaped, err := r.DB().Integration.Delete().Where(integration.ID(integrationID), integration.ExpiresAtLTE(time.Now())).Exec(ctx)
	if err != nil {
		return false, err
	}

	if reaped == 0 {
		return false, nil
	}

	return true, r.deleteCredential(ctx, integrationID)
}

// Disconnect executes the teardown flow for one installation
func (r *Runtime) Disconnect(ctx context.Context, installation *ent.Integration) (types.DisconnectResult, error) {
	def, err := r.resolveDefinitionForInstallation(installation)
	if err != nil {
		return types.DisconnectResult{}, err
	}

	connection, err := r.resolvePersistedConnection(def, installation)
	if err != nil && !errors.Is(err, ErrConnectionRequired) && !errors.Is(err, ErrConnectionNotFound) {
		return types.DisconnectResult{}, err
	}

	var result types.DisconnectResult

	if err == nil && connection.Disconnect != nil && connection.Disconnect.Disconnect != nil {
		credentials, loadErr := r.keystore().LoadCredentials(ctx, installation, connection.CredentialRefs)
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

	ctx = intobvs.WithInstallation(ctx, installation)

	skip := lo.Ternary(credential != nil, []types.CredentialSlotID{credentialRef}, nil)

	if err := r.ensureCurrentVersion(ctx, installation, skip...); err != nil {
		return err
	}

	ctx = upgradedInstallationKey.Set(ctx, installation.ID)

	wasErrored := installation.Status == enums.IntegrationStatusErrored

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

	if wasErrored {
		if err := r.recoverErroredInstallation(ctx, installation, def, credential == nil); err != nil {
			return err
		}
	}

	if credential != nil || wasErrored {
		r.assessOperationHealth(ctx, installation, def)
	}

	return nil
}

// recoverErroredInstallation clears an errored installation's unhealthy state
func (r *Runtime) recoverErroredInstallation(ctx context.Context, installation *ent.Integration, def types.Definition, verify bool) error {
	if verify {
		if err := r.verifyInstallationHealth(ctx, installation, def); err != nil {
			return err
		}
	}

	return r.ClearIntegrationUnhealthy(ctx, installation)
}

// reconcileUserInput validates and persists user input for one installation
func (r *Runtime) reconcileUserInput(ctx context.Context, installation *ent.Integration, def types.Definition, userInput json.RawMessage) error {
	if def.UserInput != nil {
		if err := validatePayload(ctx, def.UserInput.Schema, userInput, ErrUserInputInvalid); err != nil {
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

	r.keystore().InvalidateClients(installation.ID)

	return r.RefreshInstallationMetadata(ctx, installation)
}

// selfInstanceMetadata identifies an installation with no external instance by its own id
func selfInstanceMetadata(installation *ent.Integration) types.IntegrationInstallationMetadata {
	return types.IntegrationInstallationMetadata{Display: types.IntegrationInstallationIdentity{ExternalID: installation.ID}}
}

// resolveConnectionIdentity resolves installation metadata via the connection's resolver
func resolveConnectionIdentity(ctx context.Context, installation *ent.Integration, def types.Definition, connection types.ConnectionRegistration, bindings types.CredentialBindings, input json.RawMessage) (types.IntegrationInstallationMetadata, error) {
	if def.Installation == nil {
		return selfInstanceMetadata(installation), nil
	}

	metadata, ok, err := def.Installation.Resolve(ctx, types.InstallationRequest{
		Integration: installation,
		Connection:  connection,
		Credentials: bindings,
		Config:      installation.Config,
		Input:       input,
	})
	if err != nil {
		return types.IntegrationInstallationMetadata{}, err
	}

	if !ok || metadata.Display.ExternalID == "" {
		return types.IntegrationInstallationMetadata{}, ErrInstallationInstanceIDRequired
	}

	return metadata, nil
}

// checkInstallationInstanceMatch rejects metadata whose external id differs from the stored one
func checkInstallationInstanceMatch(installation *ent.Integration, metadata types.IntegrationInstallationMetadata) error {
	stored := installation.InstallationMetadata.Display.ExternalID
	resolved := metadata.Display.ExternalID

	if stored == "" || stored == resolved {
		return nil
	}

	return fmt.Errorf("%w: stored %s, resolved %s", ErrInstallationInstanceMismatch, stored, resolved)
}

// saveInstallationMetadata persists metadata and syncs display identity into the metadata map
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

// RefreshInstallationMetadata re-resolves and persists installation metadata from its credential
func (r *Runtime) RefreshInstallationMetadata(ctx context.Context, installation *ent.Integration) error {
	def, err := r.resolveDefinitionForInstallation(installation)
	if err != nil {
		return err
	}

	state, err := def.ProviderState(installation.ProviderState)
	if err != nil {
		return err
	}

	systemCtx := privacy.DecisionContext(ctx, privacy.Allow)

	if state.CredentialRef == (types.CredentialSlotID{}) {
		if len(def.Connections) > 0 {
			return nil
		}

		return r.saveInstallationMetadata(systemCtx, installation, selfInstanceMetadata(installation))
	}

	connection, err := def.ConnectionRegistration(state.CredentialRef)
	if err != nil {
		return err
	}

	bindings, err := r.loadCredentials(systemCtx, installation, connection.CredentialRefs)
	if err != nil {
		return err
	}

	metadata, err := resolveConnectionIdentity(systemCtx, installation, def, connection, bindings, nil)
	if err != nil {
		return err
	}

	metadata.Display.CredentialRef = state.CredentialRef.String()

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

	if err := validatePayload(ctx, registration.Schema, credential.Data, ErrCredentialInvalid); err != nil {
		return err
	}

	bindings, err := r.keystore().LoadCredentials(ctx, installation, connection.CredentialRefs)
	if err != nil {
		return err
	}

	bindings = bindings.With(credentialRef, credential)

	if def.HealthCheck != nil {
		if err := r.runConnectionHealthCheck(ctx, installation, def.HealthCheck, bindings); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
	}

	systemCtx := privacy.DecisionContext(ctx, privacy.Allow)

	if err := r.RefreshInstallationMetadata(systemCtx, installation); err != nil {
		logx.FromContext(systemCtx).Debug().Err(err).Msg("reconcile: instance id refresh before match check failed; comparing against stored id")
	}

	metadata, err := resolveConnectionIdentity(systemCtx, installation, def, connection, bindings, installationInput)
	if err != nil {
		return err
	}

	if err := checkInstallationInstanceMatch(installation, metadata); err != nil {
		return err
	}

	metadata.Display.CredentialRef = credentialRef.String()

	if err := r.keystore().SaveCredential(systemCtx, installation, registration.Ref, credential); err != nil {
		return err
	}

	if err := r.persistConnectionState(systemCtx, installation, def, connection.CredentialRef); err != nil {
		return err
	}

	if err := r.saveInstallationMetadata(systemCtx, installation, metadata); err != nil {
		return err
	}

	return r.activateReconciledInstallation(systemCtx, installation, def)
}

// activateReconciledInstallation records the credential and runs first-connection setup
func (r *Runtime) activateReconciledInstallation(ctx context.Context, installation *ent.Integration, def types.Definition) error {
	wasFirstConnection := installation.Status == enums.IntegrationStatusPending
	wasErrored := installation.Status == enums.IntegrationStatusErrored

	health := installation.Health
	health.UnhealthyOperations = nil
	installation.Health = health

	version := r.Registry().Version(def.ID)

	update := r.DB().Integration.UpdateOneID(installation.ID).SetHealth(health).SetDefinitionVersion(version).ClearExpiresAt()

	if !wasErrored {
		update = update.SetStatus(enums.IntegrationStatusConnected)
	}

	if err := update.Exec(ctx); err != nil {
		return err
	}

	installation.DefinitionVersion = version

	if !wasErrored {
		installation.Status = enums.IntegrationStatusConnected
	}

	if err := r.reconcileInstallationWebhooks(ctx, installation, ""); err != nil {
		return err
	}

	if !wasFirstConnection {
		return nil
	}

	if err := r.ResetReconcileLoops(ctx, installation); err != nil {
		return err
	}

	r.notifyIntegrationInstalled(ctx, installation, def)

	return nil
}

// resolveConnectionFromState resolves the connection persisted in provider state
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
		return types.ConnectionRegistration{}, false, fmt.Errorf("%w: %w", ErrConnectionNotFound, err)
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

	switch {
	case err != nil:
		return types.ConnectionRegistration{}, err
	case found && !lo.Contains(connection.CredentialRefs, credentialRef):
		return types.ConnectionRegistration{}, ErrCredentialNotDeclared
	case found:
		return connection, nil
	case credentialRef == (types.CredentialSlotID{}):
		return types.ConnectionRegistration{}, ErrConnectionRequired
	}

	connection, err = def.ConnectionRegistration(credentialRef)
	if err != nil {
		return types.ConnectionRegistration{}, fmt.Errorf("%w: %w", ErrConnectionNotFound, err)
	}

	return connection, nil
}

// notifyIntegrationInstalled dispatches a Slack message on first connection; failures are logged
func (r *Runtime) notifyIntegrationInstalled(ctx context.Context, installation *ent.Integration, def types.Definition) {
	org, err := r.DB().Organization.Get(ctx, installation.OwnerID)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to resolve organization for install notification")

		return
	}

	config, err := json.Marshal(slackdef.IntegrationInstalledMessage{
		IntegrationName:  def.DisplayName,
		OrganizationName: org.DisplayName,
		OrganizationID:   org.ID,
	})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to marshal install notification config")

		return
	}

	if _, err := r.Dispatch(ctx, types.DispatchRequest{
		DefinitionID: slackdef.DefinitionID.ID(),
		Operation:    slackdef.IntegrationInstalledOp.Name(),
		Config:       config,
		RunType:      enums.IntegrationRunTypeEvent,
		Runtime:      true,
	}); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to dispatch install notification")
	}
}

// persistConnectionState updates provider state with a new credential reference
func (r *Runtime) persistConnectionState(ctx context.Context, installation *ent.Integration, def types.Definition, credentialRef types.CredentialSlotID) error {
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
