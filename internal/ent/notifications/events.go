package notifications

import (
	"context"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/export"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/standard"
	"github.com/theopenlane/core/internal/ent/generated/task"
	"github.com/theopenlane/core/pkg/gala"
)

// orgUserIDsByRole returns the user ids of org members holding any of the given roles
func orgUserIDsByRole(ctx context.Context, client *generated.Client, orgID string, roles ...enums.Role) ([]string, error) {
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
			Fields:  []string{task.FieldAssigneeID},
			Caller:  notificationCaller,
			Notify: &entityops.NotifySpec{
				Recipients: entityops.RecipientsFromField(task.FieldAssigneeID),
				Content: entityops.NotificationContent{
					Type:          enums.NotificationTypeUser,
					Topic:         enums.NotificationTopicTaskAssignment,
					TitleTemplate: "New task assigned",
					BodyTemplate:  "Task {{ .Name }} has been assigned to you",
				},
			},
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
