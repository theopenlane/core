package notifications

import (
	"fmt"
	"strings"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/events"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/export"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/pkg/events/soiree"
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

// handleExportMutation processes export mutations and creates notifications when status changes to READY or FAILED
func handleExportMutation(ctx *soiree.EventContext, payload *events.MutationPayload) error {
	props := ctx.Properties()
	if props == nil {
		return nil
	}

	statusVal := props.GetKey(export.FieldStatus)
	if statusVal == nil {
		return nil
	}

	fields, err := fetchExportFields(ctx, props, payload)
	if err != nil {
		logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to get export fields")
		return err
	}

	// we only want to add notifications for either READY or FAILED
	if fields.status != enums.ExportStatusReady && fields.status != enums.ExportStatusFailed {
		return nil
	}

	if fields.requestorID == "" {
		logx.FromContext(ctx.Context()).Warn().Msg("requestor_id not set for export")
		return nil
	}

	if err := addExportNotification(ctx, fields); err != nil {
		logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to add export notification")
		return err
	}

	return nil
}

// fetchExportFields collects export fields from payload/properties with DB fallback.
func fetchExportFields(ctx *soiree.EventContext, props soiree.Properties, payload *events.MutationPayload) (*exportFields, error) {
	fields := &exportFields{}

	extractExportFromPayload(payload, fields)
	extractExportFromProps(props, fields)

	if needsExportDBQuery(fields) {
		if err := queryExportFromDB(ctx, fields); err != nil {
			return nil, err
		}
	}

	return fields, nil
}

// extractExportFromPayload populates export fields from mutation metadata and typed mutation values.
func extractExportFromPayload(payload *events.MutationPayload, fields *exportFields) {
	if payload == nil {
		return
	}

	if payload.EntityID != "" {
		fields.entityID = payload.EntityID
	}

	if ownerID, exists := mutationProposedString(payload, export.FieldOwnerID); exists {
		fields.ownerID = ownerID
	}

	if requestorID, exists := mutationProposedString(payload, export.FieldRequestorID); exists {
		fields.requestorID = requestorID
	}

	if exportTypeRaw, exists := mutationProposedValue(payload, export.FieldExportType); exists {
		if exportType, ok := parseExportType(exportTypeRaw); ok {
			fields.exportType = exportType
		}
	}

	if statusRaw, exists := mutationProposedValue(payload, export.FieldStatus); exists {
		if status, ok := parseExportStatus(statusRaw); ok {
			fields.status = status
		}
	}

	if errorMessage, exists := mutationProposedString(payload, export.FieldErrorMessage); exists {
		fields.errorMessage = errorMessage
	}

	exportMut, ok := payload.Mutation.(*generated.ExportMutation)
	if !ok {
		return
	}

	if ownerID, exists := exportMut.OwnerID(); exists {
		fields.ownerID = ownerID
	}

	if requestorID, exists := exportMut.RequestorID(); exists {
		fields.requestorID = requestorID
	}

	if exportType, exists := exportMut.ExportType(); exists {
		fields.exportType = exportType
	}

	if status, exists := exportMut.Status(); exists {
		fields.status = status
	}

	if errorMessage, exists := exportMut.ErrorMessage(); exists {
		fields.errorMessage = errorMessage
	}
}

// extractExportFromProps populates export fields from event properties when metadata is missing.
func extractExportFromProps(props soiree.Properties, fields *exportFields) {
	if fields.ownerID == "" {
		if ownerID, ok := props.GetKey(export.FieldOwnerID).(string); ok {
			fields.ownerID = ownerID
		}
	}

	if fields.requestorID == "" {
		if requestorID, ok := props.GetKey(export.FieldRequestorID).(string); ok {
			fields.requestorID = requestorID
		}
	}

	if fields.exportType == "" {
		if exportType, ok := parseExportType(props.GetKey(export.FieldExportType)); ok {
			fields.exportType = exportType
		}
	}

	if fields.status == "" {
		if status, ok := parseExportStatus(props.GetKey(export.FieldStatus)); ok {
			fields.status = status
		}
	}

	if fields.errorMessage == "" {
		if errorMessage, ok := props.GetKey(export.FieldErrorMessage).(string); ok {
			fields.errorMessage = errorMessage
		}
	}

	if fields.entityID == "" {
		if id, ok := props.GetKey(export.FieldID).(string); ok {
			fields.entityID = id
		}
	}
}

// parseExportType parses an export type value from metadata/property variants.
func parseExportType(raw any) (enums.ExportType, bool) {
	return events.ParseEnum(raw, enums.ToExportType, enums.ExportTypeInvalid)
}

// parseExportStatus parses an export status value from metadata/property variants.
func parseExportStatus(raw any) (enums.ExportStatus, bool) {
	return events.ParseEnum(raw, enums.ToExportStatus, enums.ExportStatusInvalid)
}

// needsExportDBQuery reports whether export fields are incomplete and require DB lookup.
func needsExportDBQuery(fields *exportFields) bool {
	return fields.entityID == "" || fields.ownerID == "" || fields.requestorID == "" || fields.exportType == ""
}

// queryExportFromDB fetches missing export fields from the database.
func queryExportFromDB(ctx *soiree.EventContext, fields *exportFields) error {
	if fields.entityID == "" {
		return ErrEntityIDNotFound
	}

	client, ok := soiree.ClientAs[*generated.Client](ctx)
	if !ok {
		return ErrFailedToGetClient
	}

	allowCtx := privacy.DecisionContext(ctx.Context(), privacy.Allow)

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

// addExportNotification creates a notification for the export requestor when eligible.
func addExportNotification(ctx *soiree.EventContext, input *exportFields) error {
	client, ok := soiree.ClientAs[*generated.Client](ctx)
	if !ok {
		return ErrFailedToGetClient
	}

	allowCtx := privacy.DecisionContext(ctx.Context(), privacy.Allow)

	// verify the requestor is a valid user (not a service account)
	// we only want to add notifications for exports coming from users not the api
	userOk, err := client.User.Query().Where(user.ID(input.requestorID)).Exist(allowCtx)
	if err != nil {
		logx.FromContext(ctx.Context()).Warn().Err(err).Msg("failed to check if requestor is a user")
		return nil
	}

	if !userOk {
		logx.FromContext(ctx.Context()).Debug().Msg("export requestor is not a user, skipping notification")
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
