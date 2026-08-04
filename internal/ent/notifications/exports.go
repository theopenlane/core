package notifications

import (
	"context"
	"fmt"
	"strings"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/export"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

type exportFields struct {
	entityID     string
	ownerID      string
	requestorID  string
	exportType   enums.ExportType
	status       enums.ExportStatus
	errorMessage string
}

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
func handleExportMutation(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	ctx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return ErrFailedToGetClient
	}

	if !eventqueue.MutationFieldChanged(payload, export.FieldStatus) {
		return nil
	}

	allowCtx := privacy.DecisionContext(ctx.Context, privacy.Allow)

	exportEntity, err := client.Export.Get(allowCtx, payload.EntityID)
	switch {
	case generated.IsNotFound(err):
		return nil
	case err != nil:
		logx.FromContext(ctx.Context).Error().Err(err).Msg("failed to query export")
		return err
	}

	// Only notify for READY or FAILED statuses.
	if exportEntity.Status != enums.ExportStatusReady && exportEntity.Status != enums.ExportStatusFailed {
		return nil
	}

	if exportEntity.RequestorID == "" {
		logx.FromContext(ctx.Context).Warn().Msg("requestor_id not set for export")
		return nil
	}

	if err := addExportNotification(ctx.Context, client, exportEntity); err != nil {
		logx.FromContext(ctx.Context).Error().Err(err).Msg("failed to add export notification")
		return err
	}

	return nil
}

// addExportNotification notifies the requesting user that their export completed or failed
func addExportNotification(ctx context.Context, client *generated.Client, exportEntity *generated.Export) error {
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	// Verify the requestor is a user (not service account) before notifying.
	userOK, err := client.User.Query().Where(user.ID(exportEntity.RequestorID)).Exist(allowCtx)
	if err != nil {
		logx.FromContext(ctx).Warn().Err(err).Msg("failed to check if requestor is a user")
		return nil
	}

	if !userOK {
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

	if _, err := client.Notification.Create().
		SetInput(*notifInput).
		SetUserID(exportEntity.RequestorID).
		Save(allowCtx); err != nil {
		return fmt.Errorf("failed to create export notification: %w", err)
	}

	return nil
}
