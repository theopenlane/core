package runtime

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/logx"
)

// integrationReconfigureObjectType is the notification object type for an installation that
// stopped syncing and needs user action to recover
const integrationReconfigureObjectType = "integration.reconfiguration.required"

// unhealthyReasonMetadataKey is the Integration.Metadata key recording why the installation
// was marked unhealthy
const unhealthyReasonMetadataKey = "unhealthyReason"

// MarkIntegrationUnhealthy flags one installation as errored with a user-facing reason and
// notifies the owning organization; recurring cycles stop on their next status check. An
// installation that is already errored is left as is so repeated failures don't stack
// duplicate notifications
func (r *Runtime) MarkIntegrationUnhealthy(ctx context.Context, installation *ent.Integration, reason string) error {
	if installation.Status == enums.IntegrationStatusErrored {
		return nil
	}

	metadata := installation.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}

	metadata[unhealthyReasonMetadataKey] = reason

	// health marking runs from worker contexts without a privileged caller; both writes are
	// server-internal and require the allow decision per the notification mutation policy
	systemCtx := privacy.DecisionContext(ctx, privacy.Allow)

	if err := r.DB().Integration.UpdateOneID(installation.ID).
		SetStatus(enums.IntegrationStatusErrored).
		SetMetadata(metadata).
		Exec(systemCtx); err != nil {
		return err
	}

	displayName := r.integrationDisplayName(installation)

	logx.FromContext(ctx).Warn().Str("reason", reason).Msg("integration marked unhealthy, recurring operations will stop")

	_, err := r.DB().Notification.Create().
		SetOwnerID(installation.OwnerID).
		SetNotificationType(enums.NotificationTypeOrganization).
		SetObjectType(integrationReconfigureObjectType).
		SetTitle(fmt.Sprintf("%s has stopped syncing", displayName)).
		SetBody(fmt.Sprintf("The %s integration has stopped syncing: %s. Reconnect it to resume.", displayName, reason)).
		SetData(map[string]any{
			"integration_id": installation.ID,
			"definition_id":  installation.DefinitionID,
			"reason":         reason,
		}).
		SetTopic(enums.NotificationTopicIntegration).
		Save(systemCtx)

	return err
}

// ClearIntegrationUnhealthy returns an errored installation to connected, removes the recorded
// reason, and reseeds its recurring operations
func (r *Runtime) ClearIntegrationUnhealthy(ctx context.Context, installation *ent.Integration) error {
	metadata := installation.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}

	delete(metadata, unhealthyReasonMetadataKey)

	if err := r.DB().Integration.UpdateOneID(installation.ID).
		SetStatus(enums.IntegrationStatusConnected).
		SetMetadata(metadata).
		Exec(privacy.DecisionContext(ctx, privacy.Allow)); err != nil {
		return err
	}

	installation.Status = enums.IntegrationStatusConnected

	return r.SeedReconcileJobsForInstallation(ctx, installation)
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
