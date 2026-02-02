package graphapi

import (
	"context"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/groupmembership"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/workflowassignmenttarget"
	"github.com/theopenlane/core/internal/graphapi/common"
	"github.com/theopenlane/core/internal/graphapi/model"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/internal/workflows/resolvers"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"
)

// AssignmentDecisionContext holds validated context for approve/reject operations
type AssignmentDecisionContext struct {
	// Assignment is the workflow assignment being decided on
	Assignment *generated.WorkflowAssignment
	// Instance is the workflow instance associated with the assignment
	Instance *generated.WorkflowInstance
	// UserID is the ID of the user making the decision
	UserID string
}

// validateAssignmentDecision performs common validation for approve/reject operations
func (r *mutationResolver) validateAssignmentDecision(ctx context.Context, id string) (*AssignmentDecisionContext, error) {
	assignment, err := withTransactionalMutation(ctx).WorkflowAssignment.Get(ctx, id)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionUpdate, Object: "workflowassignment"})
	}

	if err := r.assertAssignmentActor(ctx, assignment); err != nil {
		return nil, err
	}

	instance, err := withTransactionalMutation(ctx).WorkflowInstance.Get(ctx, assignment.WorkflowInstanceID)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionGet, Object: "workflowinstance"})
	}
	if instance.State != enums.WorkflowInstanceStatePaused {
		return nil, rout.ErrPermissionDenied
	}

	if assignment.Status != enums.WorkflowAssignmentStatusPending {
		return nil, rout.ErrPermissionDenied
	}

	userID, _ := auth.GetSubjectIDFromContext(ctx) // error already validated by assertAssignmentActor

	return &AssignmentDecisionContext{
		Assignment: assignment,
		Instance:   instance,
		UserID:     userID,
	}, nil
}

// assertAssignmentActor checks that the user in the context is a valid actor for the given assignment
func (r *mutationResolver) assertAssignmentActor(ctx context.Context, assignment *generated.WorkflowAssignment) error {
	userID, err := auth.GetSubjectIDFromContext(ctx)
	if err != nil || userID == "" {
		return rout.ErrPermissionDenied
	}

	directTarget, err := withTransactionalMutation(ctx).WorkflowAssignmentTarget.Query().
		Where(
			workflowassignmenttarget.WorkflowAssignmentIDEQ(assignment.ID),
			workflowassignmenttarget.TargetUserIDEQ(userID),
		).
		Exist(ctx)
	if err != nil {
		return rout.ErrPermissionDenied
	}
	if directTarget {
		return nil
	}

	groupIDs, err := withTransactionalMutation(ctx).GroupMembership.Query().
		Where(groupmembership.UserIDEQ(userID)).
		Select(groupmembership.FieldGroupID).
		Strings(ctx)
	if err != nil {
		logx.FromContext(ctx).Warn().Err(err).Str("user_id", userID).Msg("failed to query group memberships for assignment actor check")
	}
	if len(groupIDs) == 0 {
		return rout.ErrPermissionDenied
	}

	groupTarget, err := withTransactionalMutation(ctx).WorkflowAssignmentTarget.Query().
		Where(
			workflowassignmenttarget.WorkflowAssignmentIDEQ(assignment.ID),
			workflowassignmenttarget.TargetGroupIDIn(groupIDs...),
		).
		Exist(ctx)
	if err != nil || !groupTarget {
		return rout.ErrPermissionDenied
	}

	return nil
}

// resolveWorkflowAssignmentTargetUsers resolves the user IDs for a workflow assignment target input
func resolveWorkflowAssignmentTargetUsers(ctx context.Context, client *generated.Client, obj *workflows.Object, input *model.WorkflowAssignmentTargetInput) ([]string, string, string, error) {
	switch input.Type {
	case enums.WorkflowTargetTypeUser:
		if input.ID == nil || *input.ID == "" {
			return nil, "", "", rout.NewMissingRequiredFieldError("targets.id")
		}
		return []string{*input.ID}, "", "", nil

	case enums.WorkflowTargetTypeGroup:
		if input.ID == nil || *input.ID == "" {
			return nil, "", "", rout.NewMissingRequiredFieldError("targets.id")
		}
		userIDs, err := resolvers.ResolveGroupMembers(ctx, client, *input.ID)
		if err != nil {
			return nil, "", "", err
		}
		return lo.Uniq(userIDs), "", *input.ID, nil

	case enums.WorkflowTargetTypeRole:
		if input.ID == nil || *input.ID == "" {
			return nil, "", "", rout.NewMissingRequiredFieldError("targets.id")
		}

		role := enums.ToRole(*input.ID)
		if role == nil || *role == enums.RoleInvalid {
			return nil, "", "", rout.InvalidField("targets.id")
		}

		orgID, err := auth.GetOrganizationIDFromContext(ctx)
		if err != nil {
			return nil, "", "", err
		}

		memberships, err := client.OrgMembership.Query().
			Where(
				orgmembership.OrganizationIDEQ(orgID),
				orgmembership.RoleEQ(*role),
			).
			All(ctx)
		if err != nil {
			return nil, "", "", err
		}

		userIDs := lo.Map(memberships, func(m *generated.OrgMembership, _ int) string {
			return m.UserID
		})

		return lo.Uniq(userIDs), "", "", nil

	case enums.WorkflowTargetTypeResolver:
		if input.ResolverKey == nil || *input.ResolverKey == "" {
			return nil, "", "", rout.NewMissingRequiredFieldError("targets.resolverKey")
		}
		if obj == nil {
			return nil, "", "", rout.ErrBadRequest
		}

		resolver, ok := resolvers.Get(*input.ResolverKey)
		if !ok {
			return nil, "", "", rout.InvalidField("targets.resolverKey")
		}

		userIDs, err := resolver(ctx, client, obj)
		if err != nil {
			return nil, "", "", err
		}

		return lo.Uniq(userIDs), *input.ResolverKey, "", nil

	default:
		return nil, "", "", rout.InvalidField("targets.type")
	}
}
