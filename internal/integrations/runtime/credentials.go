package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	slackdef "github.com/theopenlane/core/v2/internal/integrations/definitions/slack"
	intobvs "github.com/theopenlane/core/v2/internal/integrations/observability"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
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

// cleanupInstallation removes credentials and the installation record for one installation
func (r *Runtime) cleanupInstallation(ctx context.Context, integrationID string) error {
	if err := r.keystore().DeleteCredential(ctx, integrationID); err != nil {
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

	return true, r.keystore().DeleteCredential(ctx, integrationID)
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
			UserInput:   installation.UserInput.Data,
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

// ReconcileCredential validates, health-checks, and stores one credential for an installation, re-derives its installation metadata, and activates it
func (r *Runtime) ReconcileCredential(ctx context.Context, installation *ent.Integration, credentialRef types.CredentialSlotID, credential types.CredentialSet, installationInput json.RawMessage) error {
	def, err := r.resolveDefinitionForInstallation(installation)
	if err != nil {
		return err
	}

	ctx = intobvs.WithInstallation(ctx, installation)

	wasErrored := installation.Status == enums.IntegrationStatusErrored

	req, records, err := r.installationRequest(ctx, installation, def)
	if err != nil {
		return err
	}

	registration, err := def.CredentialRegistration(credentialRef)
	if err != nil {
		return err
	}

	connection, err := r.resolveConnectionForCredential(def, installation, credentialRef)
	if err != nil {
		return err
	}

	if err := operations.ValidateInput(ctx, req, registration.Schema, registration.Stored.Validate, credential.Data, ErrCredentialInvalid); err != nil {
		return err
	}

	bindings := types.CredentialBindings(lo.FilterMap(connection.CredentialRefs, func(ref types.CredentialSlotID, _ int) (types.CredentialBinding, bool) {
		stored, ok := records[ref]

		return types.CredentialBinding{Ref: ref, Credential: stored}, ok
	})).With(credentialRef, credential)

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

	if err := r.ensureCurrentVersion(ctx, installation); err != nil {
		return err
	}

	if err := r.activateReconciledInstallation(auth.EnsureIntegrationCaller(ctx, installation.OwnerID), installation, def); err != nil {
		return err
	}

	if wasErrored {
		if err := r.ClearIntegrationUnhealthy(ctx, installation); err != nil {
			return err
		}
	}

	r.assessOperationHealth(ctx, installation, def)

	return nil
}

// installationRequest bundles the installation, its persisted connection, every stored credential, and stored user input for validation and upgrade hooks
func (r *Runtime) installationRequest(ctx context.Context, installation *ent.Integration, def types.Definition) (types.InstallationRequest, map[types.CredentialSlotID]types.CredentialSet, error) {
	connection, err := r.resolvePersistedConnection(def, installation)
	if err != nil && !errors.Is(err, ErrConnectionRequired) && !errors.Is(err, ErrConnectionNotFound) {
		return types.InstallationRequest{}, nil, err
	}

	records, err := r.keystore().LoadAllCredentials(ctx, installation)
	if err != nil {
		return types.InstallationRequest{}, nil, err
	}

	return types.InstallationRequest{
		Integration: installation,
		Connection:  connection,
		Credentials: lo.MapToSlice(records, func(slot types.CredentialSlotID, credential types.CredentialSet) types.CredentialBinding {
			return types.CredentialBinding{Ref: slot, Credential: credential}
		}),
		UserInput: installation.UserInput.Data,
	}, records, nil
}

// selfInstanceMetadata identifies an installation with no external instance by its own id
func selfInstanceMetadata(installationID string) types.IntegrationInstallationMetadata {
	return types.IntegrationInstallationMetadata{Display: types.IntegrationInstallationIdentity{ExternalID: installationID}}
}

// resolveConnectionIdentity resolves installation metadata via the connection's resolver
func resolveConnectionIdentity(ctx context.Context, installation *ent.Integration, def types.Definition, connection types.ConnectionRegistration, bindings types.CredentialBindings, input json.RawMessage) (types.IntegrationInstallationMetadata, error) {
	if def.Installation == nil {
		return selfInstanceMetadata(installation.ID), nil
	}

	metadata, ok, err := def.Installation.Resolve(ctx, types.InstallationRequest{
		Integration: installation,
		Connection:  connection,
		Credentials: bindings,
		UserInput:   installation.UserInput.Data,
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

// RefreshInstallationMetadata re-resolves and persists installation metadata from its persisted connection
func (r *Runtime) RefreshInstallationMetadata(ctx context.Context, installation *ent.Integration) error {
	def, err := r.resolveDefinitionForInstallation(installation)
	if err != nil {
		return err
	}

	connection, found, err := r.resolveConnectionFromState(def, installation)
	if err != nil {
		return err
	}

	if state.CredentialRef == (types.CredentialSlotID{}) {
		if len(def.Connections) > 0 {
			return nil
		}

		return r.saveInstallationMetadata(ctx, installation, selfInstanceMetadata(installation))
	}

	connection, err := def.ConnectionRegistration(state.CredentialRef)
	if err != nil {
		return err
	}

	bindings, err := r.loadCredentials(ctx, installation, connection.CredentialRefs)
	if err != nil {
		return err
	}

	metadata, err := resolveConnectionIdentity(ctx, installation, def, connection, bindings, nil)
	if err != nil {
		return err
	}

	metadata.Display.CredentialRef = connection.CredentialRef.String()

	return r.saveInstallationMetadata(ctx, installation, metadata)
}

// reconcileCredential validates, health-checks, and persists one credential for an installation
func (r *Runtime) reconcileCredential(ctx context.Context, req types.InstallationRequest, installation *ent.Integration, def types.Definition, credentialRef types.CredentialSlotID, credential types.CredentialSet, installationInput json.RawMessage) error {
	registration, err := def.CredentialRegistration(credentialRef)
	if err != nil {
		return err
	}

	connection, err := r.resolveConnectionForCredential(def, installation, credentialRef)
	if err != nil {
		return err
	}

	if err := operations.ValidateInput(ctx, req, registration.Schema, registration.Validate, credential.Data, ErrCredentialInvalid); err != nil {
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

	if err := r.RefreshInstallationMetadata(ctx, installation); err != nil {
		logx.FromContext(ctx).Debug().Err(err).Msg("reconcile: instance id refresh before match check failed; comparing against stored id")
	}

	metadata, err := resolveConnectionIdentity(ctx, installation, def, connection, bindings, installationInput)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to resolve connection identity")

		return err
	}

	if err := checkInstallationInstanceMatch(installation, metadata); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("installation instance mismatch")

		return err
	}

	metadata.Display.CredentialRef = credentialRef.String()

	if err := r.keystore().SaveCredential(ctx, installation, registration.Ref.ID(), credential); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to save credential")

		return err
	}

	if err := r.persistConnectionState(ctx, installation, def, connection.CredentialRef); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to persist connection state")

		return err
	}

	if err := r.saveInstallationMetadata(ctx, installation, metadata); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to save installation metadata")

		return err
	}

	return r.activateReconciledInstallation(ctx, installation, def)
}

// activateReconciledInstallation records the credential and runs first-connection setup
func (r *Runtime) activateReconciledInstallation(ctx context.Context, installation *ent.Integration, def types.Definition) error {
	wasFirstConnection := installation.Status == enums.IntegrationStatusPending
	wasErrored := installation.Status == enums.IntegrationStatusErrored

	health := installation.Health
	health.UnhealthyOperations = nil
	installation.Health = health

	update := r.DB().Integration.UpdateOneID(installation.ID).SetHealth(health).ClearExpiresAt()

	if !wasErrored {
		update = update.SetStatus(enums.IntegrationStatusConnected)
	}

	if err := update.Exec(ctx); err != nil {
		return err
	}

	if !wasErrored {
		installation.Status = enums.IntegrationStatusConnected
	}

	if err := r.reconcileInstallationWebhooks(ctx, r.DB(), installation, ""); err != nil {
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

	credentialRef := state.CredentialRef

	if registration, replaced, ok := def.ResolveCredential(credentialRef); ok && replaced {
		credentialRef = registration.Ref
	}

	connection, err := def.ConnectionRegistration(credentialRef)
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
