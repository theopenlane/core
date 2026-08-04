package notifications

import (
	"context"
	"fmt"

	"entgo.io/ent"
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

// documentNotificationInput carries the resolved document fields required to
// notify the approver group that a policy or procedure needs approval
type documentNotificationInput struct {
	// approverID is the group whose members receive the notification
	approverID string
	// name is the document display name used in the notification body
	name string
	// docID is the document entity ID used to build the console URL
	docID string
	// ownerID is the owning organization of the document
	ownerID string
	// objectType is the ent schema type of the document
	objectType string
	// objectLabel is the human-readable document kind used in the notification title
	objectLabel string
}

// handleTaskMutation processes task mutations and creates notifications when assignee changes or mentions are added
func handleTaskMutation(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	ctx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return ErrFailedToGetClient
	}

	if assigneeID, ok := eventqueue.MutationStringValue(payload, task.FieldAssigneeID); ok {
		allowCtx := privacy.DecisionContext(ctx.Context, privacy.Allow)

		taskEntity, err := client.Task.Get(allowCtx, payload.EntityID)
		switch {
		case generated.IsNotFound(err):
			return nil
		case err != nil:
			logx.FromContext(ctx.Context).Error().Err(err).Msg("failed to query task")
			return err
		}

		if err := addTaskAssigneeNotification(ctx.Context, client, assigneeID, taskEntity); err != nil {
			logx.FromContext(ctx.Context).Error().Err(err).Msg("failed to add task assignee notification")
			return err
		}
	}

	return handleObjectMentions(ctx, payload)
}

// handleInternalPolicyMutation processes internal policy mutations and creates notifications when status requires approval or mentions are added
func handleInternalPolicyMutation(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	if err := handleDocumentNeedsApproval(ctx, payload); err != nil {
		return err
	}

	return handleObjectMentions(ctx, payload)
}

// handleProcedureMutation processes procedure mutations and creates notifications for mentions and approval requests
func handleProcedureMutation(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	if err := handleDocumentNeedsApproval(ctx, payload); err != nil {
		return err
	}

	return handleObjectMentions(ctx, payload)
}

// handleRiskMutation processes risk mutations and creates notifications for mentions
func handleRiskMutation(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	return handleObjectMentions(ctx, payload)
}

// handleDocumentNeedsApproval notifies the approver group when a policy or procedure
// transitions to NEEDS_APPROVAL; status field names are shared between internalpolicy
// and procedure so the internalpolicy constant covers both
func handleDocumentNeedsApproval(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	rawStatus, ok := eventqueue.MutationValue(payload, internalpolicy.FieldStatus)
	if !ok {
		return nil
	}

	status, ok := eventqueue.ParseEnum(rawStatus, enums.ToDocumentStatus, enums.DocumentStatusInvalid)
	if !ok || status != enums.DocumentNeedsApproval {
		return nil
	}

	ctx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return ErrFailedToGetClient
	}

	allowCtx := privacy.DecisionContext(ctx.Context, privacy.Allow)
	input := documentNotificationInput{docID: payload.EntityID, objectType: payload.MutationType}

	switch payload.MutationType {
	case generated.TypeInternalPolicy:
		policy, err := client.InternalPolicy.Get(allowCtx, payload.EntityID)
		switch {
		case generated.IsNotFound(err):
			return nil
		case err != nil:
			logx.FromContext(ctx.Context).Error().Err(err).Msg("failed to query internal policy")
			return err
		}

		input.approverID, input.name, input.ownerID, input.objectLabel = policy.ApproverID, policy.Name, policy.OwnerID, "Internal Policy"
	case generated.TypeProcedure:
		proc, err := client.Procedure.Get(allowCtx, payload.EntityID)
		switch {
		case generated.IsNotFound(err):
			return nil
		case err != nil:
			logx.FromContext(ctx.Context).Error().Err(err).Msg("failed to query procedure")
			return err
		}

		input.approverID, input.name, input.ownerID, input.objectLabel = proc.ApproverID, proc.Name, proc.OwnerID, "Procedure"
	default:
		return nil
	}

	if input.approverID == "" {
		logx.FromContext(ctx.Context).Warn().Msg("approver_id not set for document with NEEDS_APPROVAL status")
		return nil
	}

	if err := addDocumentNotification(ctx.Context, client, input); err != nil {
		logx.FromContext(ctx.Context).Error().Err(err).Msg("failed to add document notification")
		return err
	}

	return nil
}

// addTaskAssigneeNotification notifies the new assignee that a task was assigned to them
func addTaskAssigneeNotification(ctx context.Context, client *generated.Client, assigneeID string, taskEntity *generated.Task) error {
	dataMap := map[string]any{
		"url": getURLPathForObject(taskEntity.ID, generated.TypeTask),
	}

	topic := enums.NotificationTopicTaskAssignment
	notifInput := &generated.CreateNotificationInput{
		NotificationType: enums.NotificationTypeUser,
		Title:            "New task assigned",
		Body:             fmt.Sprintf("Task %s has been assigned to you", taskEntity.Title),
		Data:             dataMap,
		OwnerID:          &taskEntity.OwnerID,
		Topic:            &topic,
		ObjectType:       generated.TypeTask,
	}

	return newNotificationCreation(ctx, client, []string{assigneeID}, notifInput)
}

// addDocumentNotification fans an approval-required notification out to every member of the approver group
func addDocumentNotification(ctx context.Context, client *generated.Client, input documentNotificationInput) error {
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

	dataMap := map[string]any{
		"url": getURLPathForObject(input.docID, input.objectType),
	}

	topic := enums.NotificationTopicApproval
	notifInput := &generated.CreateNotificationInput{
		NotificationType: enums.NotificationTypeOrganization,
		Title:            fmt.Sprintf("%s approval required", input.objectLabel),
		Body:             fmt.Sprintf("%s needs approval", input.name),
		Data:             dataMap,
		OwnerID:          &input.ownerID,
		Topic:            &topic,
		ObjectType:       input.objectType,
	}

	return newNotificationCreation(ctx, client, userIDs, notifInput)
}

// newNotificationCreation creates one notification per user for the given input
func newNotificationCreation(ctx context.Context, client *generated.Client, userIDs []string, input *generated.CreateNotificationInput) error {
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
func RegisterGalaListeners(g *gala.Gala) ([]gala.ListenerID, error) {
	return gala.Register(g,
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
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic:      eventqueue.MutationTopic(eventqueue.MutationConcernNotification, generated.TypeProgram),
			Name:       "notifications.program",
			Operations: []string{ent.OpUpdate.String(), ent.OpUpdateOne.String()},
			Handle:     handleProgramMutation,
		},
	)
}
