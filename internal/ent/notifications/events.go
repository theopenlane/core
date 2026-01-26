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
type policyFields struct {
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

type policyNotificationInput struct {
	approverID string
	policyName string
	policyID   string
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
	props := ctx.Properties()
	if props == nil {
		return nil
	}

	// Check if status field changed - only trigger notification if this field was updated
	statusVal := props.GetKey(internalpolicy.FieldStatus)
	if statusVal != nil {
		status, ok := statusVal.(enums.DocumentStatus)
		if ok {

			// Check if status is NEEDS_APPROVAL
			if status == enums.DocumentNeedsApproval {
				// Get approver_id from payload and props, fallback to database query if missing
				fields, err := fetchPolicyFields(ctx, props, payload)
				if err != nil {
					logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to get internal policy fields")
					return err
				}

				if fields.approverID == "" {
					logx.FromContext(ctx.Context()).Warn().Msg("approver_id not set for internal policy with NEEDS_APPROVAL status")
				} else {
					input := policyNotificationInput{
						approverID: fields.approverID,
						policyName: fields.name,
						policyID:   fields.entityID,
						ownerID:    fields.ownerID,
					}

					if err := addInternalPolicyNotification(ctx, input); err != nil {
						logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to add internal policy notification")
						return err
					}
				}
			}
		}
	}

	// Check for mentions in policy details
	if err := handleObjectMentions(ctx, payload); err != nil {
		logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to handle internal policy mentions")
		return err
	}

	return nil
}

// handleRiskMutation processes risk mutations and creates notifications for mentions
func handleRiskMutation(ctx *soiree.EventContext, payload *events.MutationPayload) error {
	return handleObjectMentions(ctx, payload)
}

// handleProcedureMutation processes procedure mutations and creates notifications for mentions
func handleProcedureMutation(ctx *soiree.EventContext, payload *events.MutationPayload) error {
	return handleObjectMentions(ctx, payload)
}

// fetchPolicyFields retrieves internal policy fields from payload, props, or queries database if missing
func fetchPolicyFields(ctx *soiree.EventContext, props soiree.Properties, payload *events.MutationPayload) (*policyFields, error) {
	fields := &policyFields{}

	extractPolicyFromPayload(payload, fields)
	extractPolicyFromProps(props, fields)

	if needsPolicyDBQuery(fields) {
		if err := queryPolicyFromDB(ctx, fields); err != nil {
			return nil, err
		}
	}

	return fields, nil
}

// extractPolicyFromPayload extracts policy fields from mutation payload
func extractPolicyFromPayload(payload *events.MutationPayload, fields *policyFields) {
	if payload == nil {
		return
	}

	if payload.EntityID != "" {
		fields.entityID = payload.EntityID
	}

	policyMut, ok := payload.Mutation.(*generated.InternalPolicyMutation)
	if !ok {
		return
	}

	if name, exists := policyMut.Name(); exists {
		fields.name = name
	}

	if ownerID, exists := policyMut.OwnerID(); exists {
		fields.ownerID = ownerID
	}

	if approverID, exists := policyMut.ApproverID(); exists {
		fields.approverID = approverID
	}
}

// extractPolicyFromProps extracts policy fields from properties
func extractPolicyFromProps(props soiree.Properties, fields *policyFields) {
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

// needsPolicyDBQuery checks if database query is needed
func needsPolicyDBQuery(fields *policyFields) bool {
	return fields.name == "" || fields.entityID == "" || fields.ownerID == "" || fields.approverID == ""
}

// queryPolicyFromDB queries policy from database to fill missing fields
func queryPolicyFromDB(ctx *soiree.EventContext, fields *policyFields) error {
	if fields.entityID == "" {
		return ErrEntityIDNotFound
	}

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

func addTaskAssigneeNotification(ctx *soiree.EventContext, input taskNotificationInput) error {
	client, ok := soiree.ClientAs[*generated.Client](ctx)
	if !ok {
		return nil
	}

	consoleURL := client.EntConfig.Notifications.ConsoleURL

	// create the data map with the URL
	dataMap := map[string]any{
		"url": fmt.Sprintf("%s/tasks?id=%s", consoleURL, input.taskID),
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

func addInternalPolicyNotification(ctx *soiree.EventContext, input policyNotificationInput) error {
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

	// create the data map with the URL
	dataMap := map[string]any{
		"url": fmt.Sprintf("%s/policies/%s", consoleURL, input.policyID),
	}

	topic := enums.NotificationTopicApproval
	notifInput := &generated.CreateNotificationInput{
		NotificationType: enums.NotificationTypeOrganization,
		Title:            "Policy approval required",
		Body:             fmt.Sprintf("%s needs approval, internalPolicy", input.policyName),
		Data:             dataMap,
		OwnerID:          &input.ownerID,
		Topic:            &topic,
		ObjectType:       "InternalPolicy",
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
			logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to create notification")
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
