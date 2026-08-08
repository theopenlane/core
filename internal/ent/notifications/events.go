package notifications

import (
	"context"
	"fmt"

	"entgo.io/ent"
	"github.com/stoewer/go-strcase"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/export"
	"github.com/theopenlane/core/internal/ent/generated/groupmembership"
	"github.com/theopenlane/core/internal/ent/generated/internalpolicy"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/standard"
	"github.com/theopenlane/core/internal/ent/generated/task"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
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
func handleTaskMutation(inv entityops.Invocation, payload entityops.MutationPayload) error {
	if assigneeID, ok := payload.StringValue(task.FieldAssigneeID); ok {
		allowCtx := privacy.DecisionContext(inv.Context, privacy.Allow)

		taskEntity, err := inv.Client.Task.Get(allowCtx, payload.EntityID)
		switch {
		case generated.IsNotFound(err):
			return nil
		case err != nil:
			logx.FromContext(inv.Context).Error().Err(err).Msg("failed to query task")
			return err
		}

		if err := addTaskAssigneeNotification(inv.Context, inv.Client, assigneeID, taskEntity); err != nil {
			logx.FromContext(inv.Context).Error().Err(err).Msg("failed to add task assignee notification")
			return err
		}
	}

	return handleObjectMentions(inv, payload)
}

// handleInternalPolicyMutation processes internal policy mutations and creates notifications when status requires approval or mentions are added
func handleInternalPolicyMutation(inv entityops.Invocation, payload entityops.MutationPayload) error {
	if err := handleDocumentNeedsApproval(inv, payload); err != nil {
		return err
	}

	return handleObjectMentions(inv, payload)
}

// handleProcedureMutation processes procedure mutations and creates notifications for mentions and approval requests
func handleProcedureMutation(inv entityops.Invocation, payload entityops.MutationPayload) error {
	if err := handleDocumentNeedsApproval(inv, payload); err != nil {
		return err
	}

	return handleObjectMentions(inv, payload)
}

// handleRiskMutation processes risk mutations and creates notifications for mentions
func handleRiskMutation(inv entityops.Invocation, payload entityops.MutationPayload) error {
	return handleObjectMentions(inv, payload)
}

// handleDocumentNeedsApproval notifies the approver group when a policy or procedure
// transitions to NEEDS_APPROVAL; status field names are shared between internalpolicy
// and procedure so the internalpolicy constant covers both
func handleDocumentNeedsApproval(inv entityops.Invocation, payload entityops.MutationPayload) error {
	rawStatus, ok := payload.Value(internalpolicy.FieldStatus)
	if !ok {
		return nil
	}

	status, ok := entityops.ParseEnum(rawStatus, enums.ToDocumentStatus, enums.DocumentStatusInvalid)
	if !ok || status != enums.DocumentNeedsApproval {
		return nil
	}

	input := documentNotificationInput{docID: payload.EntityID, objectType: payload.MutationType}

	switch payload.MutationType {
	case generated.TypeInternalPolicy:
		input.objectLabel = "Internal Policy"
	case generated.TypeProcedure:
		input.objectLabel = "Procedure"
	default:
		return nil
	}

	schema, ok := entityops.LookupSchema(payload.MutationType)
	if !ok {
		return nil
	}

	allowCtx := privacy.DecisionContext(inv.Context, privacy.Allow)

	row, err := schema.Load(allowCtx, inv.Client, payload.EntityID)
	switch {
	case generated.IsNotFound(err):
		return nil
	case err != nil:
		logx.FromContext(inv.Context).Error().Err(err).Msg("failed to load document")
		return err
	}

	fields, err := jsonx.Decode[map[string]any](row)
	if err != nil {
		logx.FromContext(inv.Context).Error().Err(err).Msg("failed to decode document row")
		return err
	}

	input.name = schema.DisplayValue(row)
	input.approverID, _ = fields[internalpolicy.FieldApproverID].(string)
	input.ownerID, _ = fields[internalpolicy.FieldOwnerID].(string)

	if input.approverID == "" {
		logx.FromContext(inv.Context).Warn().Msg("approver_id not set for document with NEEDS_APPROVAL status")
		return nil
	}

	if err := addDocumentNotification(inv.Context, inv.Client, input); err != nil {
		logx.FromContext(inv.Context).Error().Err(err).Msg("failed to add document notification")
		return err
	}

	return nil
}

// addTaskAssigneeNotification notifies the new assignee that a task was assigned to them
func addTaskAssigneeNotification(ctx context.Context, client *generated.Client, assigneeID string, taskEntity *generated.Task) error {
	dataMap := map[string]any{
		"url": entityops.ConsoleObjectPath(generated.TypeTask, taskEntity.ID),
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
	ctx = logx.WithFields(ctx, map[string]any{"group_id": input.approverID})
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	groupMemberships, err := client.GroupMembership.Query().
		Where(groupmembership.GroupID(input.approverID)).
		All(allowCtx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to get approver group")
		return err
	}

	if len(groupMemberships) == 0 {
		logx.FromContext(ctx).Warn().Msg("no users found in approver group")
		return nil
	}

	userIDs := make([]string, len(groupMemberships))
	for i, gm := range groupMemberships {
		userIDs[i] = gm.UserID
	}

	dataMap := map[string]any{
		"url": entityops.ConsoleObjectPath(input.objectType, input.docID),
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
		entityops.MutationListener{
			Concern: gala.MutationConcernNotification,
			Schema:  generated.TypeTask,
			Handle:  handleTaskMutation,
		}.Definition(),
		entityops.MutationListener{
			Concern: gala.MutationConcernNotification,
			Schema:  generated.TypeInternalPolicy,
			Handle:  handleInternalPolicyMutation,
		}.Definition(),
		entityops.MutationListener{
			Concern: gala.MutationConcernNotification,
			Schema:  generated.TypeRisk,
			Handle:  handleRiskMutation,
		}.Definition(),
		entityops.MutationListener{
			Concern: gala.MutationConcernNotification,
			Schema:  generated.TypeProcedure,
			Handle:  handleProcedureMutation,
		}.Definition(),
		entityops.MutationListener{
			Concern: gala.MutationConcernNotification,
			Schema:  generated.TypeNote,
			Handle:  handleNoteMutation,
		}.Definition(),
		entityops.MutationListener{
			Concern: gala.MutationConcernNotification,
			Schema:  generated.TypeExport,
			Fields:  []string{export.FieldStatus},
			Handle:  handleExportMutation,
		}.Definition(),
		entityops.MutationListener{
			Concern:    gala.MutationConcernNotification,
			Schema:     generated.TypeStandard,
			Operations: []string{ent.OpUpdate.String(), ent.OpUpdateOne.String()},
			Fields:     []string{standard.FieldRevision},
			Handle:     handleStandardMutation,
		}.Definition(),
		entityops.MutationListener{
			Concern:    gala.MutationConcernNotification,
			Schema:     generated.TypeProgram,
			Operations: []string{ent.OpUpdate.String(), ent.OpUpdateOne.String()},
			Handle:     handleProgramMutation,
		}.Definition(),
	)
}
