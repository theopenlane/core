package notifications

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/stoewer/go-strcase"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/groupmembership"
	"github.com/theopenlane/core/internal/ent/generated/internalpolicy"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/task"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

var (
	// ErrFailedToGetClient is returned when the client cannot be retrieved from context.
	ErrFailedToGetClient = errors.New("failed to get client from context")
	// ErrEntityIDNotFound is returned when an entity ID is unavailable in payload metadata.
	ErrEntityIDNotFound = errors.New("entity ID not found in payload metadata")
)

// jsonSliceToString converts a []any slice to a JSON string for parsing.
// Returns an empty string if the slice is empty or serialization fails.
func jsonSliceToString(data []any) string {
	if len(data) == 0 {
		return ""
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return ""
	}

	return string(bytes)
}

type taskFields struct {
	title    string
	entityID string
	ownerID  string
}

type documentFields struct {
	approverID string
	name       string
	entityID   string
	ownerID    string
}

type taskNotificationInput struct {
	assigneeID string
	taskTitle  string
	taskID     string
	ownerID    string
}

type documentNotificationInput struct {
	approverID string
	name       string
	docID      string
	ownerID    string
}

// handleTaskMutation processes task mutations and creates notifications when assignee changes or mentions are added.
func handleTaskMutation(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return ErrFailedToGetClient
	}

	props := ctx.Envelope.Headers.Properties

	if eventqueue.MutationFieldChanged(payload, task.FieldAssigneeID) {
		assigneeID := eventqueue.MutationStringValueOrProperty(payload, props, task.FieldAssigneeID)

		if assigneeID != "" {
			fields, err := fetchTaskFields(ctx.Context, client, props, payload)
			if err != nil {
				logx.FromContext(ctx.Context).Error().Err(err).Msg("failed to get task fields")
				return err
			}

			input := taskNotificationInput{
				assigneeID: assigneeID,
				taskTitle:  fields.title,
				taskID:     fields.entityID,
				ownerID:    fields.ownerID,
			}

			if err := addTaskAssigneeNotification(ctx.Context, client, input); err != nil {
				logx.FromContext(ctx.Context).Error().Err(err).Msg("failed to add task assignee notification")
				return err
			}
		}
	}

	// Check for mentions in task details.
	if err := handleObjectMentions(ctx, payload); err != nil {
		logx.FromContext(ctx.Context).Error().Err(err).Msg("failed to handle task mentions")
		return err
	}

	return nil
}

// fetchTaskFields retrieves task fields from payload metadata, headers, or database fallback.
func fetchTaskFields(ctx context.Context, client *generated.Client, props map[string]string, payload eventqueue.MutationGalaPayload) (*taskFields, error) {
	fields := &taskFields{}

	extractTaskFromPayload(payload, fields)
	extractTaskFromProps(props, fields)

	if needsTaskDBQuery(fields) {
		if err := queryTaskFromDB(ctx, client, fields); err != nil {
			return nil, err
		}
	}

	return fields, nil
}

// extractTaskFromPayload extracts task fields from mutation payload metadata.
func extractTaskFromPayload(payload eventqueue.MutationGalaPayload, fields *taskFields) {
	if fields == nil {
		return
	}

	if payload.EntityID != "" {
		fields.entityID = payload.EntityID
	}

	if title, ok := eventqueue.MutationStringValue(payload, task.FieldTitle); ok {
		fields.title = title
	}

	if ownerID, ok := eventqueue.MutationStringValue(payload, task.FieldOwnerID); ok {
		fields.ownerID = ownerID
	}
}

// extractTaskFromProps extracts task fields from mutation header properties.
func extractTaskFromProps(props map[string]string, fields *taskFields) {
	if fields == nil {
		return
	}

	if fields.title == "" {
		fields.title = eventqueue.MutationStringFromProperties(props, task.FieldTitle)
	}

	if fields.entityID == "" {
		fields.entityID = eventqueue.MutationStringFromProperties(props, task.FieldID)
	}

	if fields.ownerID == "" {
		fields.ownerID = eventqueue.MutationStringFromProperties(props, task.FieldOwnerID)
	}
}

// needsTaskDBQuery checks if database query is needed.
func needsTaskDBQuery(fields *taskFields) bool {
	return fields == nil || fields.title == "" || fields.entityID == "" || fields.ownerID == ""
}

// queryTaskFromDB queries task from database to fill missing fields.
func queryTaskFromDB(ctx context.Context, client *generated.Client, fields *taskFields) error {
	if fields == nil || fields.entityID == "" {
		return ErrEntityIDNotFound
	}

	if client == nil {
		return ErrFailedToGetClient
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	taskEntity, err := client.Task.Get(allowCtx, fields.entityID)
	if err != nil {
		return fmt.Errorf("failed to query task: %w", err)
	}

	if fields.title == "" {
		fields.title = taskEntity.Title
	}

	if fields.ownerID == "" {
		fields.ownerID = taskEntity.OwnerID
	}

	return nil
}

// handleInternalPolicyMutation processes internal policy mutations and creates notifications when status requires approval or mentions are added.
func handleInternalPolicyMutation(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	if err := handleDocumentNeedsApproval(ctx, payload); err != nil {
		logx.FromContext(ctx.Context).Error().Err(err).Msg("failed to handle document needs approval")
		return err
	}

	if err := handleObjectMentions(ctx, payload); err != nil {
		logx.FromContext(ctx.Context).Error().Err(err).Msg("failed to handle internal policy mentions")
		return err
	}

	return nil
}

func handleDocumentNeedsApproval(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	props := ctx.Envelope.Headers.Properties

	rawStatus, ok := eventqueue.MutationValue(payload, internalpolicy.FieldStatus)
	if !ok {
		status := eventqueue.MutationStringFromProperties(props, internalpolicy.FieldStatus)
		if status == "" {
			return nil
		}

		rawStatus = status
	}

	status, ok := eventqueue.ParseEnum(rawStatus, enums.ToDocumentStatus, enums.DocumentStatusInvalid)
	if !ok || status != enums.DocumentNeedsApproval {
		return nil
	}

	client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return ErrFailedToGetClient
	}

	fields, err := fetchDocumentFields(ctx.Context, client, props, payload)
	if err != nil {
		logx.FromContext(ctx.Context).Error().Err(err).Msg("failed to get document fields")
		return err
	}

	if fields.approverID == "" {
		logx.FromContext(ctx.Context).Warn().Msg("approver_id not set for document with NEEDS_APPROVAL status")
		return nil
	}

	input := documentNotificationInput{
		approverID: fields.approverID,
		name:       fields.name,
		docID:      fields.entityID,
		ownerID:    fields.ownerID,
	}

	if err := addDocumentNotification(ctx.Context, client, input, payload); err != nil {
		logx.FromContext(ctx.Context).Error().Err(err).Msg("failed to add document notification")
		return err
	}

	return nil
}

// handleRiskMutation processes risk mutations and creates notifications for mentions.
func handleRiskMutation(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	return handleObjectMentions(ctx, payload)
}

// handleProcedureMutation processes procedure mutations and creates notifications for mentions and approval requests.
func handleProcedureMutation(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	if err := handleDocumentNeedsApproval(ctx, payload); err != nil {
		logx.FromContext(ctx.Context).Error().Err(err).Msg("failed to handle document needs approval")
		return err
	}

	return handleObjectMentions(ctx, payload)
}

// fetchDocumentFields retrieves document fields from payload metadata, headers, or database fallback.
func fetchDocumentFields(ctx context.Context, client *generated.Client, props map[string]string, payload eventqueue.MutationGalaPayload) (*documentFields, error) {
	fields := &documentFields{}

	extractDocumentFromPayload(payload, fields)
	extractDocumentFromProps(props, fields)

	if needsDocumentDBQuery(fields) {
		if err := queryDocumentFromDB(ctx, client, fields, payload); err != nil {
			return nil, err
		}
	}

	return fields, nil
}

// extractDocumentFromPayload extracts document fields from mutation payload metadata.
func extractDocumentFromPayload(payload eventqueue.MutationGalaPayload, fields *documentFields) {
	if fields == nil {
		return
	}

	if payload.EntityID != "" {
		fields.entityID = payload.EntityID
	}

	if name, ok := eventqueue.MutationStringValue(payload, internalpolicy.FieldName); ok {
		fields.name = name
	}

	if ownerID, ok := eventqueue.MutationStringValue(payload, internalpolicy.FieldOwnerID); ok {
		fields.ownerID = ownerID
	}

	if approverID, ok := eventqueue.MutationStringValue(payload, internalpolicy.FieldApproverID); ok {
		fields.approverID = approverID
	}
}

// extractDocumentFromProps extracts document fields from header properties.
// This uses internalpolicy constants because field names are shared with procedure.
func extractDocumentFromProps(props map[string]string, fields *documentFields) {
	if fields == nil {
		return
	}

	if fields.approverID == "" {
		fields.approverID = eventqueue.MutationStringFromProperties(props, internalpolicy.FieldApproverID)
	}

	if fields.name == "" {
		fields.name = eventqueue.MutationStringFromProperties(props, internalpolicy.FieldName)
	}

	if fields.entityID == "" {
		fields.entityID = eventqueue.MutationStringFromProperties(props, internalpolicy.FieldID)
	}

	if fields.ownerID == "" {
		fields.ownerID = eventqueue.MutationStringFromProperties(props, internalpolicy.FieldOwnerID)
	}
}

// needsDocumentDBQuery checks if database query is needed.
func needsDocumentDBQuery(fields *documentFields) bool {
	return fields == nil ||
		fields.name == "" ||
		fields.entityID == "" ||
		fields.ownerID == "" ||
		fields.approverID == ""
}

// queryDocumentFromDB queries document records from database to fill missing fields.
func queryDocumentFromDB(ctx context.Context, client *generated.Client, fields *documentFields, payload eventqueue.MutationGalaPayload) error {
	if fields == nil || fields.entityID == "" {
		return ErrEntityIDNotFound
	}

	if client == nil {
		return ErrFailedToGetClient
	}

	switch payload.MutationType {
	case generated.TypeInternalPolicy:
		return getInternalPolicyFromDB(ctx, client, fields)
	case generated.TypeProcedure:
		return getProcedureFromDB(ctx, client, fields)
	default:
		logx.FromContext(ctx).Warn().Str("mutation_type", payload.MutationType).Msg("unsupported mutation type for document query")
	}

	return nil
}

// getInternalPolicyFromDB queries internal policy from database to fill missing fields.
func getInternalPolicyFromDB(ctx context.Context, client *generated.Client, fields *documentFields) error {
	if client == nil {
		return ErrFailedToGetClient
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	policy, err := client.InternalPolicy.Get(allowCtx, fields.entityID)
	if err != nil {
		return fmt.Errorf("failed to query internal policy: %w", err)
	}

	if fields.name == "" {
		fields.name = policy.Name
	}

	if fields.ownerID == "" {
		fields.ownerID = policy.OwnerID
	}

	if fields.approverID == "" && policy.ApproverID != "" {
		fields.approverID = policy.ApproverID
	}

	return nil
}

// getProcedureFromDB queries procedure from database to fill missing fields.
func getProcedureFromDB(ctx context.Context, client *generated.Client, fields *documentFields) error {
	if client == nil {
		return ErrFailedToGetClient
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	procedure, err := client.Procedure.Get(allowCtx, fields.entityID)
	if err != nil {
		return fmt.Errorf("failed to query procedure: %w", err)
	}

	if fields.name == "" {
		fields.name = procedure.Name
	}

	if fields.ownerID == "" {
		fields.ownerID = procedure.OwnerID
	}

	if fields.approverID == "" && procedure.ApproverID != "" {
		fields.approverID = procedure.ApproverID
	}

	return nil
}

func addTaskAssigneeNotification(ctx context.Context, client *generated.Client, input taskNotificationInput) error {
	if client == nil {
		return ErrFailedToGetClient
	}

	consoleURL := client.EntConfig.Notifications.ConsoleURL

	dataMap := map[string]any{
		"url": getURLPathForObject(consoleURL, input.taskID, generated.TypeTask),
	}

	topic := enums.NotificationTopicTaskAssignment
	notifInput := &generated.CreateNotificationInput{
		NotificationType: enums.NotificationTypeUser,
		Title:            "New task assigned",
		Body:             fmt.Sprintf("Task %s has been assigned to you", input.taskTitle),
		Data:             dataMap,
		OwnerID:          &input.ownerID,
		Topic:            &topic,
		ObjectType:       generated.TypeTask,
	}

	return newNotificationCreation(ctx, client, []string{input.assigneeID}, notifInput)
}

func addDocumentNotification(ctx context.Context, client *generated.Client, input documentNotificationInput, payload eventqueue.MutationGalaPayload) error {
	if client == nil {
		return ErrFailedToGetClient
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	groupMemberships, err := client.GroupMembership.Query().
		Where(groupmembership.GroupID(input.approverID)).
		All(allowCtx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("group_id", input.approverID).Msg("failed to get approver group")
		return err
	}

	if len(groupMemberships) == 0 {
		logx.FromContext(ctx).Warn().Str("group_id", input.approverID).Msg("no users found in approver group")
		return nil
	}

	userIDs := make([]string, len(groupMemberships))
	for i, gm := range groupMemberships {
		userIDs[i] = gm.UserID
	}

	consoleURL := client.EntConfig.Notifications.ConsoleURL

	objectName := ""
	switch payload.MutationType {
	case generated.TypeInternalPolicy:
		objectName = "Internal Policy"
	case generated.TypeProcedure:
		objectName = "Procedure"
	default:
		logx.FromContext(ctx).Warn().Str("mutation_type", payload.MutationType).Msg("unsupported mutation type for document notification")
		return nil
	}

	dataMap := map[string]any{
		"url": getURLPathForObject(consoleURL, input.docID, payload.MutationType),
	}

	topic := enums.NotificationTopicApproval
	notifInput := &generated.CreateNotificationInput{
		NotificationType: enums.NotificationTypeOrganization,
		Title:            fmt.Sprintf("%s approval required", objectName),
		Body:             fmt.Sprintf("%s needs approval", input.name),
		Data:             dataMap,
		OwnerID:          &input.ownerID,
		Topic:            &topic,
		ObjectType:       payload.MutationType,
	}

	return newNotificationCreation(ctx, client, userIDs, notifInput)
}

func newNotificationCreation(ctx context.Context, client *generated.Client, userIDs []string, input *generated.CreateNotificationInput) error {
	if client == nil {
		return ErrFailedToGetClient
	}

	// Ensure object type is normalized.
	input.ObjectType = strcase.UpperSnakeCase(input.ObjectType)

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	for _, userID := range userIDs {
		mut := client.Notification.Create().SetInput(*input).SetUserID(userID)
		if _, err := mut.Save(allowCtx); err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("user_id", userID).Msg("failed to create notification")
			return err
		}
	}

	return nil
}

// RegisterGalaListeners registers mutation listeners for notifications on Gala.
func RegisterGalaListeners(registry *gala.Registry) ([]gala.ListenerID, error) {
	return gala.RegisterListeners(registry,
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic:  eventqueue.MutationTopic(eventqueue.MutationConcernNotification, generated.TypeTask),
			Name:   "notifications.task",
			Handle: handleTaskMutation,
		},
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic:  eventqueue.MutationTopic(eventqueue.MutationConcernNotification, generated.TypeInternalPolicy),
			Name:   "notifications.internal_policy",
			Handle: handleInternalPolicyMutation,
		},
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic:  eventqueue.MutationTopic(eventqueue.MutationConcernNotification, generated.TypeRisk),
			Name:   "notifications.risk",
			Handle: handleRiskMutation,
		},
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic:  eventqueue.MutationTopic(eventqueue.MutationConcernNotification, generated.TypeProcedure),
			Name:   "notifications.procedure",
			Handle: handleProcedureMutation,
		},
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic:  eventqueue.MutationTopic(eventqueue.MutationConcernNotification, generated.TypeNote),
			Name:   "notifications.note",
			Handle: handleNoteMutation,
		},
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic:  eventqueue.MutationTopic(eventqueue.MutationConcernNotification, generated.TypeExport),
			Name:   "notifications.export",
			Handle: handleExportMutation,
		},
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic:  eventqueue.MutationTopic(eventqueue.MutationConcernNotification, generated.TypeStandard),
			Name:   "notifications.standard_update",
			Handle: handleStandardMutation,
		},
	)
}
