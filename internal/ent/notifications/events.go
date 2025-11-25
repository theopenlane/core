package notifications

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/groupmembership"
	"github.com/theopenlane/core/internal/ent/generated/internalpolicy"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/task"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/events/soiree"
	"github.com/theopenlane/core/pkg/logx"
)

var (
	// ErrFailedToGetClient is returned when the client cannot be retrieved from context
	ErrFailedToGetClient = errors.New("failed to get client from context")
	// ErrEntityIDNotFound is returned when entity ID is not found in props
	ErrEntityIDNotFound = errors.New("entity ID not found in props")
)

type notificationData struct {
	title            string
	body             string
	channels         []string
	userIDs          []string
	data             string
	ownerID          string
	notificationType string
	topic            string
	objectType       string
}

// handleTaskMutation processes task mutations and creates notifications when assignee changes
func handleTaskMutation(ctx *soiree.EventContext) error {
	event := ctx.Event()
	if event == nil {
		return nil
	}

	props := ctx.Properties()
	if props == nil {
		return nil
	}

	// Check if assignee_id field changed - only trigger notification if this field was updated
	assigneeIDVal := props.GetKey(task.FieldAssigneeID)
	if assigneeIDVal == nil {
		return nil
	}

	assigneeID, ok := assigneeIDVal.(string)
	if !ok || assigneeID == "" {
		return nil
	}

	// Get other fields from props, fallback to database query if missing
	title, entityID, ownerID, err := getTaskFields(ctx, props)
	if err != nil {
		logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to get task fields")
		return err
	}

	if err := addTaskAssigneeNotification(ctx, assigneeID, title, entityID, ownerID); err != nil {
		logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to add task assignee notification")
		return err
	}

	return nil
}

// getTaskFields retrieves task fields from props or queries database if missing
func getTaskFields(ctx *soiree.EventContext, props soiree.Properties) (title, entityID, ownerID string, err error) {
	// Try to get fields from props first
	if titleVal := props.GetKey(task.FieldTitle); titleVal != nil {
		if t, ok := titleVal.(string); ok {
			title = t
		}
	}

	if idVal := props.GetKey(task.FieldID); idVal != nil {
		if id, ok := idVal.(string); ok {
			entityID = id
		}
	}

	if ownerVal := props.GetKey(task.FieldOwnerID); ownerVal != nil {
		if o, ok := ownerVal.(string); ok {
			ownerID = o
		}
	}

	// If any field is missing, query the database
	if title == "" || entityID == "" || ownerID == "" {
		client, ok := soiree.ClientAs[*generated.Client](ctx)
		if !ok {
			return "", "", "", ErrFailedToGetClient
		}

		// Use the entity ID from props to query
		if entityID == "" {
			return "", "", "", ErrEntityIDNotFound
		}

		allowCtx := privacy.DecisionContext(ctx.Context(), privacy.Allow)
		taskEntity, err := client.Task.Get(allowCtx, entityID)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to query task: %w", err)
		}

		if title == "" {
			title = taskEntity.Title
		}
		if ownerID == "" {
			ownerID = taskEntity.OwnerID
		}
	}

	return title, entityID, ownerID, nil
}

// handleInternalPolicyMutation processes internal policy mutations and creates notifications when status = NEEDS_APPROVAL
func handleInternalPolicyMutation(ctx *soiree.EventContext) error {
	event := ctx.Event()
	if event == nil {
		return nil
	}

	props := ctx.Properties()
	if props == nil {
		return nil
	}

	// Check if status field changed - only trigger notification if this field was updated
	statusVal := props.GetKey(internalpolicy.FieldStatus)
	if statusVal == nil {
		return nil
	}

	status, ok := statusVal.(string)
	if !ok {
		return nil
	}

	statusEnum := enums.ToDocumentStatus(status)

	// Check if status is NEEDS_APPROVAL
	if statusEnum != &enums.DocumentNeedsApproval {
		return nil
	}

	// Get approver_id from props, fallback to database query if missing
	approverID, name, entityID, ownerID, err := getInternalPolicyFields(ctx, props)
	if err != nil {
		logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to get internal policy fields")
		return err
	}

	if approverID == "" {
		logx.FromContext(ctx.Context()).Warn().Msg("approver_id not set for internal policy with NEEDS_APPROVAL status")
		return nil
	}

	if err := addInternalPolicyNotification(ctx, approverID, name, entityID, ownerID); err != nil {
		logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to add internal policy notification")
		return err
	}

	return nil
}

// getInternalPolicyFields retrieves internal policy fields from props or queries database if missing
func getInternalPolicyFields(ctx *soiree.EventContext, props soiree.Properties) (approverID, name, entityID, ownerID string, err error) {
	// Try to get fields from props first
	if approverVal := props.GetKey(internalpolicy.FieldApproverID); approverVal != nil {
		if a, ok := approverVal.(string); ok {
			approverID = a
		}
	}

	if nameVal := props.GetKey(internalpolicy.FieldName); nameVal != nil {
		if n, ok := nameVal.(string); ok {
			name = n
		}
	}

	if idVal := props.GetKey(internalpolicy.FieldID); idVal != nil {
		if id, ok := idVal.(string); ok {
			entityID = id
		}
	}

	if ownerVal := props.GetKey(internalpolicy.FieldOwnerID); ownerVal != nil {
		if o, ok := ownerVal.(string); ok {
			ownerID = o
		}
	}

	// If any field is missing, query the database
	if name == "" || entityID == "" || ownerID == "" || approverID == "" {
		client, ok := soiree.ClientAs[*generated.Client](ctx)
		if !ok {
			return "", "", "", "", ErrFailedToGetClient
		}

		// Use the entity ID from props to query
		if entityID == "" {
			return "", "", "", "", ErrEntityIDNotFound
		}

		allowCtx := privacy.DecisionContext(ctx.Context(), privacy.Allow)
		policy, err := client.InternalPolicy.Get(allowCtx, entityID)
		if err != nil {
			return "", "", "", "", fmt.Errorf("failed to query internal policy: %w", err)
		}

		if name == "" {
			name = policy.Name
		}
		if ownerID == "" {
			ownerID = policy.OwnerID
		}
		if approverID == "" && policy.ApproverID != "" {
			approverID = policy.ApproverID
		}
	}

	return approverID, name, entityID, ownerID, nil
}

func addTaskAssigneeNotification(ctx *soiree.EventContext, assigneeID, taskTitle, taskID, ownerID string) error {
	client, ok := soiree.ClientAs[*generated.Client](ctx)
	if !ok {
		return nil
	}

	consoleURL := client.EntConfig.Notifications.ConsoleURL

	// create the data map with the URL
	dataMap := map[string]string{
		"url": fmt.Sprintf("%s/tasks?id=%s", consoleURL, taskID),
	}

	dataJSON, err := json.Marshal(dataMap)
	if err != nil {
		logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to marshal notification data")
		return err
	}

	data := notificationData{
		notificationType: enums.NotificationTypeUser.String(),
		userIDs:          []string{assigneeID},
		title:            "New task assigned",
		body:             fmt.Sprintf("Task %s has been assigned to you", taskTitle),
		data:             string(dataJSON),
		ownerID:          ownerID,
		topic:            "task_assignment",
		objectType:       "Task",
	}

	return newNotificationCreation(ctx, data)
}

func addInternalPolicyNotification(ctx *soiree.EventContext, approverID, policyName, policyID, ownerID string) error {
	client, ok := soiree.ClientAs[*generated.Client](ctx)
	if !ok {
		return nil
	}

	// set allow context to query the group
	allowCtx := privacy.DecisionContext(ctx.Context(), privacy.Allow)

	users, err :=
		client.GroupMembership.Query().Where(groupmembership.GroupID(approverID)).All(allowCtx)
	if err != nil {
		logx.FromContext(ctx.Context()).Error().Err(err).Str("group_id", approverID).Msg("failed to get approver group")
		return err
	}

	if len(users) == 0 {
		logx.FromContext(ctx.Context()).Warn().Str("group_id", approverID).Msg("no users found in approver group")
		return nil
	}

	// collect user IDs
	userIDs := make([]string, len(users))
	for i, user := range users {
		userIDs[i] = user.ID
	}

	consoleURL := client.EntConfig.Notifications.ConsoleURL

	// create the data map with the URL
	dataMap := map[string]string{
		"url": fmt.Sprintf("%s/policies/%s/view", consoleURL, policyID),
	}

	dataJSON, err := json.Marshal(dataMap)
	if err != nil {
		logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to marshal notification data")
		return err
	}

	data := notificationData{
		notificationType: enums.NotificationTypeOrganization.String(),
		userIDs:          userIDs,
		title:            "Policy approval required",
		body:             fmt.Sprintf("%s needs approval, internalPolicy", policyName),
		data:             string(dataJSON),
		ownerID:          ownerID,
		topic:            "policy_approval",
		objectType:       "InternalPolicy",
	}

	return newNotificationCreation(ctx, data)
}

func newNotificationCreation(ctx *soiree.EventContext, data notificationData) error {
	client, ok := soiree.ClientAs[*generated.Client](ctx)
	if !ok {
		return nil
	}

	// set allow context
	allowCtx := privacy.DecisionContext(ctx.Context(), privacy.Allow)

	// create notification per user that it should be sent to
	for _, userID := range data.userIDs {
		mut := client.Notification.Create()

		// Set owner ID
		if data.ownerID != "" {
			mut.SetOwnerID(data.ownerID)
		}

		mut.SetBody(data.body)
		mut.SetTitle(data.title)

		// set object type and topic
		if data.objectType != "" {
			mut.SetObjectType(data.objectType)
		}

		if data.topic != "" {
			mut.SetTopic(data.topic)
		}

		// set notification type
		if data.notificationType != "" {
			notifType := enums.ToNotificationType(data.notificationType)
			if notifType != nil && *notifType != enums.NotificationTypeInvalid {
				mut.SetNotificationType(*notifType)
			}
		}

		// set data if provided
		if data.data != "" {
			var dataMap map[string]interface{}
			if err := json.Unmarshal([]byte(data.data), &dataMap); err != nil {
				logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to unmarshal notification data")
			} else {
				mut.SetData(dataMap)
			}
		}

		// set channels if provided
		if len(data.channels) > 0 {
			// convert string channels to enums.Channel and set them
			channels := make([]enums.Channel, 0, len(data.channels))
			for _, ch := range data.channels {
				if c := enums.ToChannel(ch); c != nil && *c != enums.ChannelInvalid {
					channels = append(channels, *c)
				}
			}
			if len(channels) > 0 {
				mut.SetChannels(channels)
			}
		}

		mut.SetUserID(userID)

		if _, err := mut.Save(allowCtx); err != nil {
			logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to create notification")
			return err
		}
	}

	return nil
}

// RegisterListeners registers notification listeners with the given eventer
// This is called from hooks package to register the listeners
func RegisterListeners(addListener func(entityType string, handler func(*soiree.EventContext) error)) {
	addListener(generated.TypeTask, handleTaskMutation)
	addListener(generated.TypeInternalPolicy, handleInternalPolicyMutation)
}
