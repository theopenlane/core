package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
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
		credential, _, loadErr := r.keystore().LoadCredential(ctx, installation, connection.Credential.Name)
		if loadErr != nil {
			return types.DisconnectResult{}, loadErr
		}

		result, err = connection.Disconnect.Disconnect(ctx, types.ConnectionInput{
			Integration: installation,
			Credential:  credential,
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
func (r *Runtime) ReconcileCredential(ctx context.Context, installation *ent.Integration, connectionName string, credential types.CredentialSet) error {
	def, err := r.resolveDefinitionForInstallation(installation)
	if err != nil {
		return err
	}

	ctx = intobvs.WithInstallation(ctx, installation)

	wasErrored := installation.Status == enums.IntegrationStatusErrored

	req, records, err := r.installationRequest(ctx, installation)
	if err != nil {
		return err
	}

	connection, ok := def.Connection(connectionName)
	if !ok {
		return ErrConnectionNotFound
	}

	persisted, found, err := r.resolveConnectionFromState(def, installation)
	if err != nil {
		return err
	}

	if found && persisted.Credential.Name != connection.Credential.Name {
		return ErrConnectionMismatch
	}

	if err := operations.ValidateInput(ctx, req, connection.Form, connection.Credential.Validate, credential.Data, ErrCredentialInvalid); err != nil {
		return err
	}

	if stored, ok := records[connection.Credential.Name]; ok {
		refreshed, refreshErr := r.runConnectionHealthCheck(ctx, installation, def, connection, stored)

		switch {
		case refreshErr != nil:
			logx.FromContext(ctx).Debug().Err(refreshErr).Msg("reconcile: identity refresh under the stored credential failed; comparing against stored id")
		default:
			if err := r.saveInstallationMetadata(ctx, installation, def, connection, refreshed); err != nil {
				return err
			}
		}
	}

	metadata, err := r.runConnectionHealthCheck(ctx, installation, def, connection, credential)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := checkInstallationInstanceMatch(installation, metadata); err != nil {
		return err
	}

	if err := r.keystore().SaveCredential(ctx, installation, connection.Credential.Name, credential); err != nil {
		return err
	}

	if err := r.persistConnectionState(ctx, installation, def, connection.Credential.Name); err != nil {
		return err
	}

	if err := r.saveInstallationMetadata(ctx, installation, def, connection, metadata); err != nil {
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

// installationRequest bundles the installation, every stored credential, and stored user input for validation and upgrade hooks
func (r *Runtime) installationRequest(ctx context.Context, installation *ent.Integration) (types.InstallationRequest, map[string]types.CredentialSet, error) {
	records, err := r.keystore().LoadAllCredentials(ctx, installation)
	if err != nil {
		return types.InstallationRequest{}, nil, err
	}

	return types.InstallationRequest{
		Integration: installation,
		Credentials: records,
		UserInput:   installation.UserInput.Data,
	}, records, nil
}

// selfInstanceMetadata identifies an installation with no external instance by its own id
func selfInstanceMetadata(installationID string) types.IntegrationInstallationMetadata {
	return types.IntegrationInstallationMetadata{Display: types.IntegrationInstallationIdentity{ExternalID: installationID}}
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

// saveInstallationMetadata stamps the layout and connection name, persists metadata, and syncs display identity into the metadata map
func (r *Runtime) saveInstallationMetadata(ctx context.Context, installation *ent.Integration, def types.Definition, connection types.Connection, metadata types.IntegrationInstallationMetadata) error {
	if def.Installation != nil {
		metadata.Layout = def.Installation.Name
	}

	metadata.Display.CredentialRef = connection.Credential.Name

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
func (r *Runtime) resolveConnectionFromState(def types.Definition, installation *ent.Integration) (types.Connection, bool, error) {
	state, err := def.ProviderState(installation.ProviderState)
	if err != nil {
		return types.Connection{}, false, err
	}

	if state.CredentialRef == "" {
		return types.Connection{}, false, nil
	}

	connection, _, ok := def.ResolveConnection(state.CredentialRef)
	if !ok {
		return types.Connection{}, false, fmt.Errorf("%w: %s", ErrConnectionNotFound, state.CredentialRef)
	}

	return connection, true, nil
}

// resolvePersistedConnection resolves the persisted connection for an installation
func (r *Runtime) resolvePersistedConnection(def types.Definition, installation *ent.Integration) (types.Connection, error) {
	connection, found, err := r.resolveConnectionFromState(def, installation)
	if err != nil {
		return types.Connection{}, err
	}

	connections := def.ConnectionList()

	switch {
	case found:
		return connection, nil
	case len(connections) == 1:
		return connections[0], nil
	}

	return types.Connection{}, ErrConnectionRequired
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

// persistConnectionState updates provider state with the active connection name
func (r *Runtime) persistConnectionState(ctx context.Context, installation *ent.Integration, def types.Definition, connectionName string) error {
	next, err := def.WithProviderState(installation.ProviderState, types.DefinitionProviderState{
		CredentialRef: connectionName,
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
