package graphapi

import (
	"context"
	"strings"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/groupmembership"
	"github.com/theopenlane/core/internal/ent/generated/workflowassignmenttarget"
	"github.com/theopenlane/core/internal/graphapi/common"
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

// resolveAssignmentActionKey derives the action key for an assignment
func resolveAssignmentActionKey(assignment *generated.WorkflowAssignment) string {
	if assignment == nil {
		return ""
	}

	if assignment.ApprovalMetadata.ActionKey != "" {
		return assignment.ApprovalMetadata.ActionKey
	}

	key := assignment.AssignmentKey
	if key == "" {
		return ""
	}

	prefixes := []string{"approval_", "review_"}
	for _, prefix := range prefixes {
		if !strings.HasPrefix(key, prefix) {
			continue
		}

		trimmed := strings.TrimPrefix(key, prefix)
		if trimmed == "" {
			return ""
		}

		parts := strings.Split(trimmed, "_")
		if len(parts) <= 1 {
			return trimmed
		}

		return strings.Join(parts[:len(parts)-1], "_")
	}

	return ""
}
