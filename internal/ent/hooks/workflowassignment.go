package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/group"
	"github.com/theopenlane/core/internal/ent/generated/groupmembership"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/workflowassignmenttarget"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
)

// HookWorkflowAssignmentDecisionAuth ensures only assignment targets can approve/reject.
func HookWorkflowAssignmentDecisionAuth() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.WorkflowAssignmentFunc(func(ctx context.Context, m *generated.WorkflowAssignmentMutation) (generated.Value, error) {
			if !workflowEngineEnabled(ctx, m.Client()) {
				return next.Mutate(ctx, m)
			}

			if rule.IsInternalRequest(ctx) {
				return next.Mutate(ctx, m)
			}

			if _, allow := privacy.DecisionFromContext(ctx); allow {
				return next.Mutate(ctx, m)
			}

			status, ok := m.Status()
			if !ok {
				return next.Mutate(ctx, m)
			}

			if status != enums.WorkflowAssignmentStatusApproved &&
				status != enums.WorkflowAssignmentStatusRejected &&
				status != enums.WorkflowAssignmentStatusChangesRequested {
				return next.Mutate(ctx, m)
			}

			assignmentID, ok := getSingleMutationID(ctx, m)
			if !ok {
				return next.Mutate(ctx, m)
			}

			userID, err := auth.GetSubjectIDFromContext(ctx)
			if err != nil {
				return nil, privacy.Denyf("workflow assignment approval requires a user")
			}

			orgID, err := auth.GetOrganizationIDFromContext(ctx)
			if err != nil {
				return nil, err
			}

			if actorID, ok := m.ActorUserID(); ok && actorID != "" && actorID != userID {
				return nil, privacy.Denyf("workflow assignment actor mismatch")
			}

			client := m.Client()
			if client == nil {
				return nil, privacy.Denyf("workflow assignment missing client")
			}

			directTarget, err := client.WorkflowAssignmentTarget.Query().
				Where(
					workflowassignmenttarget.WorkflowAssignmentIDEQ(assignmentID),
					workflowassignmenttarget.TargetUserIDEQ(userID),
					workflowassignmenttarget.OwnerIDEQ(orgID),
				).
				Exist(ctx)
			if err != nil {
				return nil, err
			}
			if directTarget {
				return next.Mutate(ctx, m)
			}

			groupIDs, err := client.GroupMembership.Query().
				Where(
					groupmembership.UserIDEQ(userID),
					groupmembership.HasGroupWith(group.OwnerIDEQ(orgID)),
				).
				Select(groupmembership.FieldGroupID).
				Strings(ctx)
			if err != nil {
				return nil, privacy.Denyf("workflow assignment actor group lookup failed")
			}

			if len(groupIDs) == 0 {
				return nil, privacy.Denyf("workflow assignment approval requires target membership")
			}

			groupTarget, err := client.WorkflowAssignmentTarget.Query().
				Where(
					workflowassignmenttarget.WorkflowAssignmentIDEQ(assignmentID),
					workflowassignmenttarget.TargetGroupIDIn(groupIDs...),
					workflowassignmenttarget.OwnerIDEQ(orgID),
				).
				Exist(ctx)
			if err != nil {
				return nil, err
			}
			if !groupTarget {
				return nil, privacy.Denyf("workflow assignment approval requires target membership")
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpUpdate|ent.OpUpdateOne)
}
