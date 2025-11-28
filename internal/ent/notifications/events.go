package notifications

import (
	"errors"
	"fmt"

	"entgo.io/ent"

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

// mutationPayload mirrors the MutationPayload struct from hooks package
// to avoid import cycle
type mutationPayload struct {
	Mutation  ent.Mutation
	Operation string
	EntityID  string
	Client    *generated.Client
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

// handleTaskMutation processes task mutations and creates notifications when assignee changes
func handleTaskMutation(ctx *soiree.EventContext, payload any) error {
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

	// Get other fields from props and payload, fallback to database query if missing
	fields, err := getTaskFields(ctx, props, payload)
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

	return nil
}



// getTaskFields retrieves task fields from payload, props, or queries database if missing
func getTaskFields(ctx *soiree.EventContext, props soiree.Properties, payload any) (*taskFields, error) {
	var fields taskFields

	// First, try to get fields from the mutation payload
	if payload != nil {
		if mutPayload, ok := payload.(*mutationPayload); ok {
			// Get entity ID from payload
			if mutPayload.EntityID != "" {
				fields.entityID = mutPayload.EntityID
			}

			// Try to get fields from the mutation if available
			if taskMut, ok := mutPayload.Mutation.(*generated.TaskMutation); ok {
				if title, exists := taskMut.Title(); exists {
					fields.title = title
				}
				if ownerID, exists := taskMut.OwnerID(); exists {
					fields.ownerID = ownerID
				}
			}
		}
	}

	// Try to get fields from props if not found in payload
	if fields.title == "" {
		if titleVal := props.GetKey(task.FieldTitle); titleVal != nil {
			if t, ok := titleVal.(string); ok {
				fields.title = t
			}
		}
	}

	if fields.entityID == "" {
		if idVal := props.GetKey(task.FieldID); idVal != nil {
			if id, ok := idVal.(string); ok {
				fields.entityID = id
			}
		}
	}

	if fields.ownerID == "" {
		if ownerVal := props.GetKey(task.FieldOwnerID); ownerVal != nil {
			if o, ok := ownerVal.(string); ok {
				fields.ownerID = o
			}
		}
	}

	// If any field is missing, query the database
	if fields.title == "" || fields.entityID == "" || fields.ownerID == "" {
		client, ok := soiree.ClientAs[*generated.Client](ctx)
		if !ok {
			return nil, ErrFailedToGetClient
		}

		// Use the entity ID from props to query
		if fields.entityID == "" {
			return nil, ErrEntityIDNotFound
		}

		allowCtx := privacy.DecisionContext(ctx.Context(), privacy.Allow)
		taskEntity, err := client.Task.Get(allowCtx, fields.entityID)
		if err != nil {
			return nil, fmt.Errorf("failed to query task: %w", err)
		}

		if fields.title == "" {
			fields.title = taskEntity.Title
		}
		if fields.ownerID == "" {
			fields.ownerID = taskEntity.OwnerID
		}
	}

	return &fields, nil
}

// handleInternalPolicyMutation processes internal policy mutations and creates notifications when status = NEEDS_APPROVAL
func handleInternalPolicyMutation(ctx *soiree.EventContext, payload any) error {
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

	// Get approver_id from payload and props, fallback to database query if missing
	fields, err := getInternalPolicyFields(ctx, props, payload)
	if err != nil {
		logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to get internal policy fields")
		return err
	}

	if fields.approverID == "" {
		logx.FromContext(ctx.Context()).Warn().Msg("approver_id not set for internal policy with NEEDS_APPROVAL status")
		return nil
	}

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

	return nil
}

// getInternalPolicyFields retrieves internal policy fields from payload, props, or queries database if missing
func getInternalPolicyFields(ctx *soiree.EventContext, props soiree.Properties, payload any) (*policyFields, error) {
	var fields policyFields

	// First, try to get fields from the mutation payload
	if payload != nil {
		if mutPayload, ok := payload.(*mutationPayload); ok {
			// Get entity ID from payload
			if mutPayload.EntityID != "" {
				fields.entityID = mutPayload.EntityID
			}

			// Try to get fields from the mutation if available
			if policyMut, ok := mutPayload.Mutation.(*generated.InternalPolicyMutation); ok {
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
		}
	}

	// Try to get fields from props if not found in payload
	if fields.approverID == "" {
		if approverVal := props.GetKey(internalpolicy.FieldApproverID); approverVal != nil {
			if a, ok := approverVal.(string); ok {
				fields.approverID = a
			}
		}
	}

	if fields.name == "" {
		if nameVal := props.GetKey(internalpolicy.FieldName); nameVal != nil {
			if n, ok := nameVal.(string); ok {
				fields.name = n
			}
		}
	}

	if fields.entityID == "" {
		if idVal := props.GetKey(internalpolicy.FieldID); idVal != nil {
			if id, ok := idVal.(string); ok {
				fields.entityID = id
			}
		}
	}

	if fields.ownerID == "" {
		if ownerVal := props.GetKey(internalpolicy.FieldOwnerID); ownerVal != nil {
			if o, ok := ownerVal.(string); ok {
				fields.ownerID = o
			}
		}
	}

	// If any field is missing, query the database
	if fields.name == "" || fields.entityID == "" || fields.ownerID == "" || fields.approverID == "" {
		client, ok := soiree.ClientAs[*generated.Client](ctx)
		if !ok {
			return nil, ErrFailedToGetClient
		}

		// Use the entity ID from props to query
		if fields.entityID == "" {
			return nil, ErrEntityIDNotFound
		}

		allowCtx := privacy.DecisionContext(ctx.Context(), privacy.Allow)
		policy, err := client.InternalPolicy.Get(allowCtx, fields.entityID)
		if err != nil {
			return nil, fmt.Errorf("failed to query internal policy: %w", err)
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
	}

	return &fields, nil
}

func addTaskAssigneeNotification(ctx *soiree.EventContext, input taskNotificationInput) error {
	client, ok := soiree.ClientAs[*generated.Client](ctx)
	if !ok {
		return nil
	}

	consoleURL := client.EntConfig.Notifications.ConsoleURL

	// create the data map with the URL
	dataMap := map[string]interface{}{
		"url": fmt.Sprintf("%s/tasks?id=%s", consoleURL, input.taskID),
	}

	topic := "task_assignment"
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

	users, err :=
		client.GroupMembership.Query().Where(groupmembership.GroupID(input.approverID)).All(allowCtx)
	if err != nil {
		logx.FromContext(ctx.Context()).Error().Err(err).Str("group_id", input.approverID).Msg("failed to get approver group")
		return err
	}

	if len(users) == 0 {
		logx.FromContext(ctx.Context()).Warn().Str("group_id", input.approverID).Msg("no users found in approver group")
		return nil
	}

	// collect user IDs
	userIDs := make([]string, len(users))
	for i, user := range users {
		userIDs[i] = user.ID
	}

	consoleURL := client.EntConfig.Notifications.ConsoleURL

	// create the data map with the URL
	dataMap := map[string]interface{}{
		"url": fmt.Sprintf("%s/policies/%s/view", consoleURL, input.policyID),
	}

	topic := "policy_approval"
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
func RegisterListeners(addListener func(entityType string, handler func(*soiree.EventContext, any) error)) {
	addListener(generated.TypeTask, handleTaskMutation)
	addListener(generated.TypeInternalPolicy, handleInternalPolicyMutation)
}
