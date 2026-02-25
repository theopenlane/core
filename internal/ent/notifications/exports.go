package notifications

import (
	"context"
	"fmt"
	"strings"

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

// handleExportMutation processes export mutations and creates notifications when status changes to READY or FAILED.
func handleExportMutation(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	ctx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return ErrFailedToGetClient
	}

	props := ctx.Envelope.Headers.Properties
	if !eventqueue.MutationFieldChanged(payload, export.FieldStatus) && eventqueue.MutationStringFromProperties(props, export.FieldStatus) == "" {
		return nil
	}

	fields, err := fetchExportFields(ctx.Context, client, props, payload)
	if err != nil {
		logx.FromContext(ctx.Context).Error().Err(err).Msg("failed to get export fields")
		return err
	}

	// Only notify for READY or FAILED statuses.
	if fields.status != enums.ExportStatusReady && fields.status != enums.ExportStatusFailed {
		return nil
	}

	if fields.requestorID == "" {
		logx.FromContext(ctx.Context).Warn().Msg("requestor_id not set for export")
		return nil
	}

	if err := addExportNotification(ctx.Context, client, fields); err != nil {
		logx.FromContext(ctx.Context).Error().Err(err).Msg("failed to add export notification")
		return err
	}

	return nil
}

func fetchExportFields(ctx context.Context, client *generated.Client, props map[string]string, payload eventqueue.MutationGalaPayload) (*exportFields, error) {
	fields := &exportFields{}

	extractExportFromPayload(payload, fields)
	extractExportFromProps(props, fields)

	if needsExportDBQuery(fields) {
		if err := queryExportFromDB(ctx, client, fields); err != nil {
			return nil, err
		}
	}

	return fields, nil
}

func extractExportFromPayload(payload eventqueue.MutationGalaPayload, fields *exportFields) {
	if fields == nil {
		return
	}

	if payload.EntityID != "" {
		fields.entityID = payload.EntityID
	}

	if ownerID, ok := eventqueue.MutationStringValue(payload, export.FieldOwnerID); ok {
		fields.ownerID = ownerID
	}

	if requestorID, ok := eventqueue.MutationStringValue(payload, export.FieldRequestorID); ok {
		fields.requestorID = requestorID
	}

	if exportType, ok := eventqueue.ParseEnum(payload.ProposedChanges[export.FieldExportType], enums.ToExportType, enums.ExportTypeInvalid); ok {
		fields.exportType = exportType
	}

	if status, ok := eventqueue.ParseEnum(payload.ProposedChanges[export.FieldStatus], enums.ToExportStatus, enums.ExportStatusInvalid); ok {
		fields.status = status
	}

	if errorMessage, ok := eventqueue.MutationStringValue(payload, export.FieldErrorMessage); ok {
		fields.errorMessage = errorMessage
	}
}

func extractExportFromProps(props map[string]string, fields *exportFields) {
	if fields == nil {
		return
	}

	if fields.ownerID == "" {
		fields.ownerID = eventqueue.MutationStringFromProperties(props, export.FieldOwnerID)
	}

	if fields.requestorID == "" {
		fields.requestorID = eventqueue.MutationStringFromProperties(props, export.FieldRequestorID)
	}

	if fields.exportType == "" {
		if exportType, ok := eventqueue.ParseEnum(props[export.FieldExportType], enums.ToExportType, enums.ExportTypeInvalid); ok {
			fields.exportType = exportType
		}
	}

	if fields.status == "" {
		if status, ok := eventqueue.ParseEnum(props[export.FieldStatus], enums.ToExportStatus, enums.ExportStatusInvalid); ok {
			fields.status = status
		}
	}

	if fields.errorMessage == "" {
		fields.errorMessage = eventqueue.MutationStringFromProperties(props, export.FieldErrorMessage)
	}

	if fields.entityID == "" {
		fields.entityID = eventqueue.MutationStringFromProperties(props, export.FieldID)
	}
}

func needsExportDBQuery(fields *exportFields) bool {
	return fields == nil ||
		fields.entityID == "" ||
		fields.ownerID == "" ||
		fields.requestorID == "" ||
		fields.exportType == ""
}

func queryExportFromDB(ctx context.Context, client *generated.Client, fields *exportFields) error {
	if fields == nil || fields.entityID == "" {
		return ErrEntityIDNotFound
	}

	if client == nil {
		return ErrFailedToGetClient
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	exportEntity, err := client.Export.Get(allowCtx, fields.entityID)
	if err != nil {
		return fmt.Errorf("failed to query export: %w", err)
	}

	if fields.ownerID == "" {
		fields.ownerID = exportEntity.OwnerID
	}

	if fields.requestorID == "" {
		fields.requestorID = exportEntity.RequestorID
	}

	if fields.exportType == "" {
		fields.exportType = exportEntity.ExportType
	}

	if fields.status == "" {
		fields.status = exportEntity.Status
	}

	if fields.errorMessage == "" {
		fields.errorMessage = exportEntity.ErrorMessage
	}

	return nil
}

func addExportNotification(ctx context.Context, client *generated.Client, input *exportFields) error {
	if client == nil {
		return ErrFailedToGetClient
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	// Verify the requestor is a user (not service account) before notifying.
	userOK, err := client.User.Query().Where(user.ID(input.requestorID)).Exist(allowCtx)
	if err != nil {
		logx.FromContext(ctx).Warn().Err(err).Msg("failed to check if requestor is a user")
		return nil
	}

	if !userOK {
		logx.FromContext(ctx).Debug().Msg("export requestor is not a user, skipping notification")
		return nil
	}

	dataMap := map[string]any{
		"export_id":   input.entityID,
		"export_type": input.exportType.String(),
	}

	var title, body string

	et := strings.ReplaceAll(strings.ToLower(input.exportType.String()), "_", " ")
	if input.status == enums.ExportStatusReady {
		title = "Export Complete"
		body = fmt.Sprintf("Export of %s is ready for download", et)
	} else {
		title = "Export Failed"
		body = fmt.Sprintf("Export of %s completed with errors", et)

		if input.errorMessage != "" {
			dataMap["errors"] = input.errorMessage
		}
	}

	topic := enums.NotificationTopicExport
	notifInput := &generated.CreateNotificationInput{
		NotificationType: enums.NotificationTypeUser,
		Title:            title,
		Body:             body,
		Data:             dataMap,
		OwnerID:          &input.ownerID,
		Topic:            &topic,
		ObjectType:       input.exportType.String(),
	}

	if _, err := client.Notification.Create().
		SetInput(*notifInput).
		SetUserID(input.requestorID).
		Save(allowCtx); err != nil {
		return fmt.Errorf("failed to create export notification: %w", err)
	}

	return nil
}
