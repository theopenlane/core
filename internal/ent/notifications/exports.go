package notifications

import (
	"context"
	"fmt"
	"strings"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/pkg/logx"
)

// canProcessNotification checks if we can process a notification for a user or a virtual support user
func canProcessNotification(ctx context.Context, client *generated.Client, requestorID string) bool {
	if requestorID == auth.SupportSubjectID {
		return true
	}

	ok, err := client.User.Query().Where(user.ID(requestorID)).Exist(ctx)
	if err != nil {
		logx.FromContext(ctx).Warn().Err(err).Msg("failed to check if requestor is a user")
		return false
	}

	return ok
}

// handleExportMutation processes export mutations and creates notifications when status changes to READY or FAILED.
func handleExportMutation(inv entityops.Invocation, payload entityops.MutationPayload) error {
	exportEntity, found, err := entityops.LoadEntity(inv.Context, payload.EntityID, inv.Client.Export.Get)
	switch {
	case err != nil:
		logx.FromContext(inv.Context).Error().Err(err).Msg("failed to query export")
		return err
	case !found:
		return nil
	}

	// Only notify for READY or FAILED statuses.
	if exportEntity.Status != enums.ExportStatusReady && exportEntity.Status != enums.ExportStatusFailed {
		return nil
	}

	if exportEntity.RequestorID == "" {
		logx.FromContext(inv.Context).Warn().Msg("requestor_id not set for export")
		return nil
	}

	if err := addExportNotification(inv.Context, inv.Client, exportEntity); err != nil {
		logx.FromContext(inv.Context).Error().Err(err).Msg("failed to add export notification")
		return err
	}

	return nil
}

// addExportNotification notifies the requesting user that their export completed or failed
func addExportNotification(ctx context.Context, client *generated.Client, exportEntity *generated.Export) error {
	// Verify the requestor is a user or the virtual support user before notifying.
	if !canProcessNotification(ctx, client, exportEntity.RequestorID) {
		logx.FromContext(ctx).Debug().Msg("export requestor is not a user, skipping notification")
		return nil
	}

	dataMap := map[string]any{
		"export_id":   exportEntity.ID,
		"export_type": exportEntity.ExportType.String(),
	}

	var title, body string

	et := strings.ReplaceAll(strings.ToLower(exportEntity.ExportType.String()), "_", " ")
	if exportEntity.Status == enums.ExportStatusReady {
		title = "Export Complete"
		body = fmt.Sprintf("Export of %s is ready for download", et)
	} else {
		title = "Export Failed"
		body = fmt.Sprintf("Export of %s completed with errors", et)

		if exportEntity.ErrorMessage != "" {
			dataMap["errors"] = exportEntity.ErrorMessage
		}
	}

	topic := enums.NotificationTopicExport
	notifInput := &generated.CreateNotificationInput{
		NotificationType: enums.NotificationTypeUser,
		Title:            title,
		Body:             body,
		Data:             dataMap,
		OwnerID:          &exportEntity.OwnerID,
		Topic:            &topic,
		ObjectType:       exportEntity.ExportType.String(),
	}

	return newNotificationCreation(ctx, client, []string{exportEntity.RequestorID}, notifInput)
}
