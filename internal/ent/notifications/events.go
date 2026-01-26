package notifications

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/events"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/groupmembership"
	"github.com/theopenlane/core/internal/ent/generated/internalpolicy"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/task"
	"github.com/theopenlane/core/pkg/events/soiree"
	"github.com/theopenlane/core/pkg/logx"
)

var (
	// ErrFailedToGetClient is returned when the client cannot be retrieved from context
	ErrFailedToGetClient = errors.New("failed to get client from context")
	// ErrEntityIDNotFound is returned when entity ID is not found in props
	ErrEntityIDNotFound = errors.New("entity ID not found in props")
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

// handleTaskMutation processes task mutations and creates notifications when assignee changes or mentions are added
func handleTaskMutation(ctx *soiree.EventContext, payload *events.MutationPayload) error {
	props := ctx.Properties()
	if props == nil {
		return nil
	}

	// Check if assignee_id field changed - only trigger notification if this field was updated
	assigneeIDVal := props.GetKey(task.FieldAssigneeID)
	if assigneeIDVal != nil {
		assigneeID, ok := assigneeIDVal.(string)
		if ok && assigneeID != "" {
			// Get other fields from props and payload, fallback to database query if missing
			fields, err := fetchTaskFields(ctx, props, payload)
			if err != nil {
				logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to get task fields")
				return err
			}

			input := taskNotificationInput{
				assigneeID: assigneeID,
				taskTitle:  fields.title,
				taskID:     fields.entityID,
				ownerID:    fields.ownerID,
			}

			if err := addTaskAssigneeNotification(ctx, input); err != nil {
				logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to add task assignee notification")
				return err
			}
		}
	}

	// Check for mentions in task details
	if err := handleObjectMentions(ctx, payload); err != nil {
		logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to handle task mentions")
		return err
	}

	return nil
}

// fetchTaskFields retrieves task fields from payload, props, or queries database if missing
func fetchTaskFields(ctx *soiree.EventContext, props soiree.Properties, payload *events.MutationPayload) (*taskFields, error) {
	fields := &taskFields{}

	extractTaskFromPayload(payload, fields)
	extractTaskFromProps(props, fields)

	if needsTaskDBQuery(fields) {
		if err := queryTaskFromDB(ctx, fields); err != nil {
			return nil, err
		}
	}

	return fields, nil
}

// extractTaskFromPayload extracts task fields from mutation payload
func extractTaskFromPayload(payload *events.MutationPayload, fields *taskFields) {
	if payload == nil {
		return
	}

	if payload.EntityID != "" {
		fields.entityID = payload.EntityID
	}

	taskMut, ok := payload.Mutation.(*generated.TaskMutation)
	if !ok {
		return
	}

	if title, exists := taskMut.Title(); exists {
		fields.title = title
	}

	if ownerID, exists := taskMut.OwnerID(); exists {
		fields.ownerID = ownerID
	}
}

// extractTaskFromProps extracts task fields from properties
func extractTaskFromProps(props soiree.Properties, fields *taskFields) {
	if fields.title == "" {
		if title, ok := props.GetKey(task.FieldTitle).(string); ok {
			fields.title = title
		}
	}

	if fields.entityID == "" {
		if id, ok := props.GetKey(task.FieldID).(string); ok {
			fields.entityID = id
		}
	}

	if fields.ownerID == "" {
		if ownerID, ok := props.GetKey(task.FieldOwnerID).(string); ok {
			fields.ownerID = ownerID
		}
	}
}

// needsTaskDBQuery checks if database query is needed
func needsTaskDBQuery(fields *taskFields) bool {
	return fields.title == "" || fields.entityID == "" || fields.ownerID == ""
}

// queryTaskFromDB queries task from database to fill missing fields
func queryTaskFromDB(ctx *soiree.EventContext, fields *taskFields) error {
	if fields.entityID == "" {
		return ErrEntityIDNotFound
	}

	client, ok := soiree.ClientAs[*generated.Client](ctx)
	if !ok {
		return ErrFailedToGetClient
	}

	allowCtx := privacy.DecisionContext(ctx.Context(), privacy.Allow)
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

// handleInternalPolicyMutation processes internal policy mutations and creates notifications when status = NEEDS_APPROVAL or mentions are added
func handleInternalPolicyMutation(ctx *soiree.EventContext, payload *events.MutationPayload) error {
	if err := handleDocumentNeedsApproval(ctx, payload); err != nil {
		logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to handle document needs approval")
		return err
	}

	// Check for mentions in policy details
	if err := handleObjectMentions(ctx, payload); err != nil {
		logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to handle internal policy mentions")
		return err
	}

	return nil
}

func handleDocumentNeedsApproval(ctx *soiree.EventContext, payload *events.MutationPayload) error {
	props := ctx.Properties()
	if props == nil {
		return nil
	}

	// Check if status field changed - only trigger notification if this field was updated
	statusVal := props.GetKey("status")
	if statusVal != nil {
		status, ok := statusVal.(enums.DocumentStatus)
		if ok {

			// Check if status is NEEDS_APPROVAL
			if status == enums.DocumentNeedsApproval {
				// Get approver_id from payload and props, fallback to database query if missing
				fields, err := fetchDocumentFields(ctx, props, payload)
				if err != nil {
					logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to get document fields")
					return err
				}

				if fields.approverID == "" {
					logx.FromContext(ctx.Context()).Warn().Msg("approver_id not set for document with NEEDS_APPROVAL status")
				} else {
					input := documentNotificationInput{
						approverID: fields.approverID,
						name:       fields.name,
						docID:      fields.entityID,
						ownerID:    fields.ownerID,
					}

					if err := addDocumentNotification(ctx, input, payload); err != nil {
						logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to add document notification")
						return err
					}
				}
			}
		}
	}
	return nil
}

// handleRiskMutation processes risk mutations and creates notifications for mentions
func handleRiskMutation(ctx *soiree.EventContext, payload *events.MutationPayload) error {
	return handleObjectMentions(ctx, payload)
}

// handleProcedureMutation processes procedure mutations and creates notifications for mentions
func handleProcedureMutation(ctx *soiree.EventContext, payload *events.MutationPayload) error {
	if err := handleDocumentNeedsApproval(ctx, payload); err != nil {
		logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to handle document needs approval")
		return err
	}

	return handleObjectMentions(ctx, payload)
}

// fetchDocumentFields retrieves document (policy, procedure, etc) fields from payload, props, or queries database if missing
func fetchDocumentFields(ctx *soiree.EventContext, props soiree.Properties, payload *events.MutationPayload) (*documentFields, error) {
	fields := &documentFields{}

	extractDocumentFromPayload(payload, fields)
	extractDocumentFromProps(props, fields)

	if needsDocumentDBQuery(fields) {
		if err := queryDocumentFromDB(ctx, fields, payload); err != nil {
			return nil, err
		}
	}

	return fields, nil
}

// approverMutation is an interface for mutations that have approver fields
type approverMutation interface {
	Name() (r string, exists bool)
	ApproverID() (r string, exists bool)
	OwnerID() (r string, exists bool)
}

// extractDocumentFromPayload extracts document fields from mutation payload
func extractDocumentFromPayload(payload *events.MutationPayload, fields *documentFields) {
	if payload == nil {
		return
	}

	if payload.EntityID != "" {
		fields.entityID = payload.EntityID
	}

	mut, ok := payload.Mutation.(approverMutation)
	if !ok {
		return
	}

	if name, exists := mut.Name(); exists {
		fields.name = name
	}

	if ownerID, exists := mut.OwnerID(); exists {
		fields.ownerID = ownerID
	}

	if approverID, exists := mut.ApproverID(); exists {
		fields.approverID = approverID
	}
}

// extractDocumentFromProps extracts document fields from properties
// this function uses the internalpolicy fields as they are common between policies and procedures
// and its better than writing the field names manually
func extractDocumentFromProps(props soiree.Properties, fields *documentFields) {
	if fields.approverID == "" {
		if approverID, ok := props.GetKey(internalpolicy.FieldApproverID).(string); ok {
			fields.approverID = approverID
		}
	}

	if fields.name == "" {
		if name, ok := props.GetKey(internalpolicy.FieldName).(string); ok {
			fields.name = name
		}
	}

	if fields.entityID == "" {
		if id, ok := props.GetKey(internalpolicy.FieldID).(string); ok {
			fields.entityID = id
		}
	}

	if fields.ownerID == "" {
		if ownerID, ok := props.GetKey(internalpolicy.FieldOwnerID).(string); ok {
			fields.ownerID = ownerID
		}
	}
}

// needsDocumentDBQuery checks if database query is needed
func needsDocumentDBQuery(fields *documentFields) bool {
	return fields.name == "" || fields.entityID == "" || fields.ownerID == "" || fields.approverID == ""
}

// queryDocumentFromDB queries policy from database to fill missing fields
func queryDocumentFromDB(ctx *soiree.EventContext, fields *documentFields, payload *events.MutationPayload) error {
	if fields.entityID == "" {
		return ErrEntityIDNotFound
	}

	switch t := payload.Mutation.(type) {
	case *generated.InternalPolicyMutation:
		return getInternalPolicyFromDB(ctx, fields)
	case *generated.ProcedureMutation:
		return getProcedureFromDB(ctx, fields)
	default:
		logx.FromContext(ctx.Context()).Warn().Msgf("unsupported mutation type %T for document query", t)
	}

	return nil
}

// getInternalPolicyFromDB queries internal policy from database to fill missing fields
func getInternalPolicyFromDB(ctx *soiree.EventContext, fields *documentFields) error {
	client, ok := soiree.ClientAs[*generated.Client](ctx)
	if !ok {
		return ErrFailedToGetClient
	}

	allowCtx := privacy.DecisionContext(ctx.Context(), privacy.Allow)
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

// getProcedureFromDB queries procedure from database to fill missing fields
func getProcedureFromDB(ctx *soiree.EventContext, fields *documentFields) error {
	client, ok := soiree.ClientAs[*generated.Client](ctx)
	if !ok {
		return ErrFailedToGetClient
	}

	allowCtx := privacy.DecisionContext(ctx.Context(), privacy.Allow)
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

func addTaskAssigneeNotification(ctx *soiree.EventContext, input taskNotificationInput) error {
	client, ok := soiree.ClientAs[*generated.Client](ctx)
	if !ok {
		return nil
	}

	consoleURL := client.EntConfig.Notifications.ConsoleURL

	// create the data map with the URL
	dataMap := map[string]any{
		"url": getURLPathForObject(consoleURL, input.taskID, "Task"),
	}

	topic := enums.NotificationTopicTaskAssignment
	notifInput := &generated.CreateNotificationInput{
		NotificationType: enums.NotificationTypeUser,
		Title:            "New task assigned",
		Body:             fmt.Sprintf("Task %s has been assigned to you", input.taskTitle),
		Data:             dataMap,
		OwnerID:          &input.ownerID,
		Topic:            &topic,
		ObjectType:       "Task",
	}

	return newNotificationCreation(ctx, []string{input.assigneeID}, notifInput)
}

func addDocumentNotification(ctx *soiree.EventContext, input documentNotificationInput, payload *events.MutationPayload) error {
	client, ok := soiree.ClientAs[*generated.Client](ctx)
	if !ok {
		return nil
	}

	// set allow context to query the group
	allowCtx := privacy.DecisionContext(ctx.Context(), privacy.Allow)

	groupMemberships, err :=
		client.GroupMembership.Query().Where(groupmembership.GroupID(input.approverID)).All(allowCtx)
	if err != nil {
		logx.FromContext(ctx.Context()).Error().Err(err).Str("group_id", input.approverID).Msg("failed to get approver group")
		return err
	}

	if len(groupMemberships) == 0 {
		logx.FromContext(ctx.Context()).Warn().Str("group_id", input.approverID).Msg("no users found in approver group")
		return nil
	}

	// collect user IDs
	userIDs := make([]string, len(groupMemberships))
	for i, gm := range groupMemberships {
		userIDs[i] = gm.UserID
	}

	consoleURL := client.EntConfig.Notifications.ConsoleURL

	objectName := ""
	switch payload.Mutation.(type) {
	case *generated.InternalPolicyMutation:
		objectName = "Internal Policy"
	case *generated.ProcedureMutation:
		objectName = "Procedure"
	default:
		logx.FromContext(ctx.Context()).Warn().Msg("unsupported mutation type for document notification")
		return nil
	}

	// create the data map with the URL
	dataMap := map[string]any{
		"url": getURLPathForObject(consoleURL, input.docID, payload.Mutation.Type()),
	}

	topic := enums.NotificationTopicApproval
	notifInput := &generated.CreateNotificationInput{
		NotificationType: enums.NotificationTypeOrganization,
		Title:            fmt.Sprintf("%s approval required", objectName),
		Body:             fmt.Sprintf("%s needs approval", input.name),
		Data:             dataMap,
		OwnerID:          &input.ownerID,
		Topic:            &topic,
		ObjectType:       payload.Mutation.Type(),
	}

	return newNotificationCreation(ctx, userIDs, notifInput)
}

func newNotificationCreation(ctx *soiree.EventContext, userIDs []string, input *generated.CreateNotificationInput) error {
	client, ok := soiree.ClientAs[*generated.Client](ctx)
	if !ok {
		return nil
	}

	// set allow context
	allowCtx := privacy.DecisionContext(ctx.Context(), privacy.Allow)

	// create notification per user that it should be sent to
	for _, userID := range userIDs {
		mut := client.Notification.Create().SetInput(*input).SetUserID(userID)

		if _, err := mut.Save(allowCtx); err != nil {
			logx.FromContext(ctx.Context()).Error().Err(err).Str("user_id", userID).Msg("failed to create notification")
			return err
		}
	}

	return nil
}

// RegisterListeners registers notification listeners with the given eventer
// This is called from hooks package to register the listeners
func RegisterListeners(addListener func(entityType string, handler func(*soiree.EventContext, *events.MutationPayload) error)) {
	addListener(generated.TypeTask, handleTaskMutation)
	addListener(generated.TypeInternalPolicy, handleInternalPolicyMutation)
	addListener(generated.TypeRisk, handleRiskMutation)
	addListener(generated.TypeProcedure, handleProcedureMutation)
	addListener(generated.TypeNote, handleNoteMutation)
	addListener(generated.TypeExport, handleExportMutation)
}
