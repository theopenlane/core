package runtime

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/entityops"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/mapx"
)

// unhealthyReasonMetadataKey is the Integration.Metadata key recording why the installation
// was marked unhealthy
const unhealthyReasonMetadataKey = "unhealthyReason"

// integrationUnhealthyObjectType is the notification object type for a stopped installation
const integrationUnhealthyObjectType = "INTEGRATION_RECONFIGURATION_REQUIRED"

// integrationHealthyObjectType is the notification object type for a recovered installation
const integrationHealthyObjectType = "INTEGRATION_RECONNECTED"

// MarkIntegrationUnhealthy flags one installation as errored with a user-facing reason and
// notifies the owning organization; recurring cycles stop on their next status check. An
// installation that is already errored is left as is so repeated failures don't stack
// duplicate notifications
func (r *Runtime) MarkIntegrationUnhealthy(ctx context.Context, installation *ent.Integration, reason string) error {
	metadata := mapx.DeepMergeMapAny(installation.Metadata, map[string]any{unhealthyReasonMetadataKey: reason})

	// health marking runs from worker contexts without a privileged caller; both writes are
	// server-internal and require the allow decision per the notification mutation policy
	systemCtx := privacy.DecisionContext(ctx, privacy.Allow)

	transitioned, err := r.DB().Integration.Update().
		Where(integration.ID(installation.ID), integration.StatusNEQ(enums.IntegrationStatusErrored)).
		SetStatus(enums.IntegrationStatusErrored).
		SetMetadata(metadata).
		Save(systemCtx)
	if err != nil {
		return err
	}

	if transitioned == 0 {
		return nil
	}

	displayName := r.integrationDisplayName(installation)

	logx.FromContext(ctx).Warn().Str("reason", reason).Msg("integration marked unhealthy, recurring operations will stop")

	return r.DB().Notification.Create().
		SetOwnerID(installation.OwnerID).
		SetNotificationType(enums.NotificationTypeOrganization).
		SetObjectType(integrationUnhealthyObjectType).
		SetTitle(fmt.Sprintf("%s has stopped syncing", displayName)).
		SetBody(fmt.Sprintf("The %s integration has stopped syncing: %s. Reconnect it to resume.", displayName, reason)).
		SetData(map[string]any{
			"integration_id": installation.ID,
			"definition_id":  installation.DefinitionID,
			"reason":         reason,
			// the console addresses integrations by definition id
			"url": entityops.ConsoleObjectPath(ent.TypeIntegration, installation.DefinitionID),
		}).
		SetTopic(enums.NotificationTopicIntegration).
		Exec(systemCtx)
}

// ClearIntegrationUnhealthy returns an errored installation to connected, notifies the
// owning organization, and reseeds its recurring operations; a non-errored installation
// is left as is so concurrent recoveries don't stack duplicate notifications
func (r *Runtime) ClearIntegrationUnhealthy(ctx context.Context, installation *ent.Integration) error {
	metadata := installation.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}

	delete(metadata, unhealthyReasonMetadataKey)

	systemCtx := privacy.DecisionContext(ctx, privacy.Allow)

	transitioned, err := r.DB().Integration.Update().
		Where(integration.ID(installation.ID), integration.StatusEQ(enums.IntegrationStatusErrored)).
		SetStatus(enums.IntegrationStatusConnected).
		SetMetadata(metadata).
		Save(systemCtx)
	if err != nil {
		return err
	}

	if transitioned == 0 {
		return nil
	}

	installation.Status = enums.IntegrationStatusConnected

	displayName := r.integrationDisplayName(installation)

	logx.FromContext(ctx).Info().Msg("integration recovered, recurring operations resume")

	if err := r.DB().Notification.Create().
		SetOwnerID(installation.OwnerID).
		SetNotificationType(enums.NotificationTypeOrganization).
		SetObjectType(integrationHealthyObjectType).
		SetTitle(fmt.Sprintf("%s is syncing again", displayName)).
		SetBody(fmt.Sprintf("The %s integration reconnected and syncing has resumed.", displayName)).
		SetData(map[string]any{
			"integration_id": installation.ID,
			"definition_id":  installation.DefinitionID,
			"url":            entityops.ConsoleObjectPath(ent.TypeIntegration, installation.DefinitionID),
		}).
		SetTopic(enums.NotificationTopicIntegration).
		Exec(systemCtx); err != nil {
		return err
	}

	return r.SeedReconcileJobsForInstallation(ctx, installation)
}

// verifyInstallationHealth runs the persisted connection's validation operation under stored
// credentials; connections without one pass
func (r *Runtime) verifyInstallationHealth(ctx context.Context, installation *ent.Integration, def types.Definition) error {
	connection, err := r.resolvePersistedConnection(def, installation)
	if err != nil {
		return err
	}

	if connection.ValidationOperation == "" {
		return nil
	}

	validationOp, err := r.Registry().Operation(def.ID, connection.ValidationOperation)
	if err != nil {
		return err
	}

	bindings, err := r.loadCredentials(privacy.DecisionContext(ctx, privacy.Allow), installation, connection.CredentialRefs)
	if err != nil {
		return err
	}

	if _, err := r.ExecuteOperation(ctx, installation, validationOp, bindings, nil); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("health check failed after configuration update, installation stays errored")

		return err
	}

	return nil
}

// IntegrationUnhealthyReason returns the recorded reason an installation was marked unhealthy,
// empty when it is healthy
func IntegrationUnhealthyReason(installation *ent.Integration) string {
	reason, _ := installation.Metadata[unhealthyReasonMetadataKey].(string)

	return reason
}

// integrationDisplayName resolves the definition display name for one installation, falling
// back to the installation's own name when the definition is no longer registered
func (r *Runtime) integrationDisplayName(installation *ent.Integration) string {
	if def, ok := r.Registry().Definition(installation.DefinitionID); ok {
		return def.DisplayName
	}

	return installation.Name
}
