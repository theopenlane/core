package notifications

import (
	"context"
	"fmt"

	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/export"
	"github.com/theopenlane/core/v2/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/v2/internal/ent/generated/standard"
	"github.com/theopenlane/core/v2/internal/ent/generated/task"
	"github.com/theopenlane/core/v2/pkg/gala"
)

// OrgUserIDsByRole returns the user ids of org members holding any of the given roles
func OrgUserIDsByRole(ctx context.Context, client *generated.Client, orgID string, roles ...enums.Role) ([]string, error) {
	var ids []string

	err := client.OrgMembership.Query().
		Where(
			orgmembership.OrganizationIDEQ(orgID),
			orgmembership.RoleIn(roles...),
		).
		Select(orgmembership.FieldUserID).
		Scan(ctx, &ids)
	if err != nil {
		return nil, err
	}

	return ids, nil
}

// notificationCaller grants the internal-operation capability so notification listeners
// pass privacy without per-query allow contexts
func notificationCaller(restored *auth.Caller, _ entityops.MutationPayload) *auth.Caller {
	return restored.WithCapabilities(auth.CapInternalOperation)
}

// Listeners creates user and organization notifications from entity mutations; mention and
// approval listeners fan out over registry schemas instead of per-schema declarations
func Listeners() []gala.Registration {
	regs := entityops.MentionListeners(entityops.MutationConcernNotification, notificationCaller, entityops.NotificationContent{
		Type:          enums.NotificationTypeUser,
		Topic:         enums.NotificationTopicMention,
		TitleTemplate: "You were mentioned",
		BodyTemplate:  "You were mentioned in {{ .Label }}: {{ .Name }}",
		Data: map[string]any{
			"object_type": "{{ .ObjectType }}",
			"object_id":   "{{ .EntityID }}",
			"object_name": "{{ .Name }}",
		},
	}, entityops.SchemaNote)

	regs = append(regs, entityops.ApprovalListeners(entityops.MutationConcernNotification, notificationCaller, entityops.NotificationContent{
		Type:          enums.NotificationTypeOrganization,
		Topic:         enums.NotificationTopicApproval,
		TitleTemplate: "{{ .Label }} approval required",
		BodyTemplate:  "{{ .Name }} needs approval",
	}, []string{string(enums.DocumentNeedsApproval)})...)

	return append(regs,
		entityops.MutationListener{
			Concern: entityops.MutationConcernNotification,
			Schema:  entityops.SchemaTask,
			Caller:  notificationCaller,
			Handle:  handleTaskAssignmentMutation,
		},
		entityops.MutationListener{
			Concern: entityops.MutationConcernNotification,
			Schema:  entityops.SchemaNote,
			Caller:  notificationCaller,
			Handle:  handleNoteMutation,
		},
		entityops.MutationListener{
			Concern: entityops.MutationConcernNotification,
			Schema:  entityops.SchemaExport,
			Fields:  []string{export.FieldStatus},
			Caller:  notificationCaller,
			Handle:  handleExportMutation,
		},
		entityops.MutationListener{
			Concern:    entityops.MutationConcernNotification,
			Schema:     entityops.SchemaStandard,
			Operations: []string{entityops.OpUpdate, entityops.OpUpdateOne},
			Fields:     []string{standard.FieldRevision},
			Caller:     notificationCaller,
			Handle:     handleStandardMutation,
		},
		entityops.MutationListener{
			Concern:    entityops.MutationConcernNotification,
			Schema:     entityops.SchemaProgram,
			Operations: []string{entityops.OpUpdate, entityops.OpUpdateOne},
			Caller:     notificationCaller,
			Handle:     handleProgramMutation,
		},
	)
}

func handleTaskAssignmentMutation(inv entityops.Invocation, payload entityops.MutationPayload) error {
	// we only care for sending notifications if assignment changes
	if !payload.FieldChanged(task.FieldAssigneeID) {
		return nil
	}

	assignee, ok := payload.StringValue(task.FieldAssigneeID)
	oldAssignee, oldExists := payload.OldStringValue(task.FieldAssigneeID)

	if ok == oldExists && assignee == oldAssignee {
		return nil
	}

	task, found, err := entityops.LoadEntity(inv.Context, payload.EntityID, inv.Client.Task.Get)
	if err != nil || !found {
		return err
	}

	if task.Status == enums.TaskStatusCompleted || task.Status == enums.TaskStatusWontDo || task.AssigneeID == "" {
		return nil
	}

	data := map[string]any{}
	if url := entityops.ConsoleObjectPath(generated.TypeTask, task.ID); url != "" {
		data["url"] = url
	}

	return entityops.CreateNotifications(inv.Context, inv.Client, []string{task.AssigneeID}, &generated.CreateNotificationInput{
		NotificationType: enums.NotificationTypeUser,
		Title:            "New task assigned",
		Body:             fmt.Sprintf("Task %s has been assigned to you", task.Title),
		Data:             data,
		Topic:            lo.ToPtr(enums.NotificationTopicTaskAssignment),
		ObjectType:       generated.TypeTask,
		OwnerID:          &task.OwnerID,
	})
}
