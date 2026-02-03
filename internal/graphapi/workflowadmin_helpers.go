package graphapi

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/workflowassignment"
	"github.com/theopenlane/core/internal/graphapi/common"
	"github.com/theopenlane/core/internal/workflows"
)

type workflowAdminCompletionDetails struct {
	InstanceID         string                      `json:"instance_id"`
	State              enums.WorkflowInstanceState `json:"state"`
	ObjectID           string                      `json:"object_id,omitempty"`
	ObjectType         enums.WorkflowObjectType    `json:"object_type,omitempty"`
	WorkflowProposalID string                      `json:"workflow_proposal_id,omitempty"`
	ForcedByUserID     string                      `json:"forced_by_user_id,omitempty"`
	AppliedProposal    bool                        `json:"applied_proposal,omitempty"`
	Reason             string                      `json:"reason,omitempty"`
	Manual             bool                        `json:"manual,omitempty"`
}

func recordWorkflowInstanceAdminCompletion(ctx context.Context, client *generated.Client, instance *generated.WorkflowInstance, forcedByUserID string, appliedProposal bool, reason *string) error {
	if client == nil || instance == nil {
		return nil
	}

	objectType, objectID, _ := workflowInstanceObjectContext(ctx, client, instance)

	details := workflowAdminCompletionDetails{
		InstanceID:         instance.ID,
		State:              instance.State,
		ObjectID:           objectID,
		ObjectType:         objectType,
		WorkflowProposalID: instance.WorkflowProposalID,
		ForcedByUserID:     forcedByUserID,
		AppliedProposal:    appliedProposal,
		Manual:             true,
	}
	if reason != nil {
		details.Reason = *reason
	}

	encoded, err := json.Marshal(details)
	if err != nil {
		return err
	}

	payload := models.WorkflowEventPayload{
		EventType: enums.WorkflowEventTypeInstanceCompleted,
		Details:   encoded,
	}

	ownerID, err := workflows.ResolveOwnerID(ctx, instance.OwnerID)
	if err != nil {
		return err
	}

	allowCtx := workflows.AllowContext(ctx)
	return client.WorkflowEvent.Create().
		SetWorkflowInstanceID(instance.ID).
		SetEventType(enums.WorkflowEventTypeInstanceCompleted).
		SetPayload(payload).
		SetOwnerID(ownerID).
		Exec(allowCtx)
}

// requireWorkflowAdmin checks that the user in the context is an admin for the given organization
func (r *Resolver) requireWorkflowAdmin(ctx context.Context, ownerID string) error {
	if ownerID == "" {
		return fmt.Errorf("%w: missing organization", rout.ErrBadRequest)
	}

	if auth.IsSystemAdminFromContext(ctx) {
		return nil
	}

	userID, err := auth.GetSubjectIDFromContext(ctx)
	if err != nil || userID == "" {
		return rout.ErrPermissionDenied
	}

	au, err := auth.GetAuthenticatedUserFromContext(ctx)
	if err != nil {
		return err
	}

	// make sure the `ownerID` matches the organgaizationID
	if au.OrganizationID != ownerID {
		return rout.ErrPermissionDenied
	}

	if au.OrganizationRole != auth.OwnerRole && au.OrganizationRole != auth.AdminRole {
		return rout.ErrPermissionDenied
	}

	return nil
}

func closeWorkflowAssignments(ctx context.Context, client *generated.Client, instanceID string, ownerID string, status enums.WorkflowAssignmentStatus, actorUserID string, reason *string) error {
	if client == nil || instanceID == "" {
		return nil
	}

	if status != enums.WorkflowAssignmentStatusApproved && status != enums.WorkflowAssignmentStatusRejected {
		return nil
	}

	allowCtx := workflows.AllowContext(ctx)
	if ownerID != "" {
		if err := common.SetOrganizationInAuthContext(allowCtx, &ownerID); err != nil {
			return err
		}
	}

	assignments, err := client.WorkflowAssignment.Query().
		Where(
			workflowassignment.WorkflowInstanceIDEQ(instanceID),
			workflowassignment.StatusEQ(enums.WorkflowAssignmentStatusPending),
		).
		All(allowCtx)
	if err != nil {
		return parseRequestError(ctx, err, common.Action{Action: common.ActionGet, Object: "workflowassignment"})
	}
	if len(assignments) == 0 {
		return nil
	}

	now := time.Now().UTC()
	skipCtx := workflows.WithSkipEventEmission(allowCtx)
	workflows.MarkSkipEventEmission(skipCtx)

	for _, assignment := range assignments {
		update := client.WorkflowAssignment.UpdateOneID(assignment.ID).
			SetStatus(status).
			SetDecidedAt(now)

		if actorUserID != "" {
			update.SetActorUserID(actorUserID)
		}

		switch status {
		case enums.WorkflowAssignmentStatusApproved:
			approvalMeta := assignment.ApprovalMetadata
			approvalMeta.ApprovedAt = now.Format(time.RFC3339)
			approvalMeta.ApprovedByUserID = actorUserID
			update.SetApprovalMetadata(approvalMeta)
		case enums.WorkflowAssignmentStatusRejected:
			rejectionMeta := assignment.RejectionMetadata
			rejectionMeta.RejectedAt = now.Format(time.RFC3339)
			rejectionMeta.RejectedByUserID = actorUserID
			if reason != nil {
				rejectionMeta.RejectionReason = *reason
			}
			update.SetRejectionMetadata(rejectionMeta)
		}

		if err := update.Exec(skipCtx); err != nil {
			return parseRequestError(ctx, err, common.Action{Action: common.ActionUpdate, Object: "workflowassignment"})
		}
	}

	return nil
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

func (r *mutationResolver) forceCompleteWorkflowInstance(ctx context.Context, id string, applyProposal bool) (*generated.WorkflowInstance, error) {
	allowCtx := workflows.AllowContext(ctx)
	instance, err := r.db.WorkflowInstance.Get(allowCtx, id)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionGet, Object: "workflowinstance"})
	}

	if err := r.requireWorkflowAdmin(ctx, instance.OwnerID); err != nil {
		return nil, err
	}

	skipCtx := workflows.WithSkipEventEmission(allowCtx)
	workflows.MarkSkipEventEmission(skipCtx)

	if instance.OwnerID != "" {
		if err := common.SetOrganizationInAuthContext(skipCtx, &instance.OwnerID); err != nil {
			return nil, err
		}
	}

	userID, _ := auth.GetSubjectIDFromContext(ctx)
	applied := false

	if instance.WorkflowProposalID != "" {
		if applyProposal {
			objectType, objectID, err := workflowInstanceObjectContext(ctx, r.db, instance)
			if err != nil {
				return nil, err
			}
			if objectID == "" || objectType == "" {
				return nil, fmt.Errorf("%w: workflow instance missing object context", rout.ErrBadRequest)
			}

			proposal, err := r.db.WorkflowProposal.Get(skipCtx, instance.WorkflowProposalID)
			if err != nil {
				return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionGet, Object: "workflowproposal"})
			}

			bypassCtx := workflows.AllowBypassContext(ctx)
			bypassCtx = workflows.WithSkipEventEmission(bypassCtx)
			workflows.MarkSkipEventEmission(bypassCtx)

			if err := workflows.ApplyObjectFieldUpdates(bypassCtx, r.db, objectType, objectID, proposal.Changes); err != nil {
				return nil, err
			}

			if err := r.db.WorkflowProposal.UpdateOneID(proposal.ID).
				SetState(enums.WorkflowProposalStateApplied).
				SetApprovedHash(proposal.ProposedHash).
				Exec(skipCtx); err != nil {
				return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionUpdate, Object: "workflowproposal"})
			}

			applied = true
		} else {
			if err := r.db.WorkflowProposal.UpdateOneID(instance.WorkflowProposalID).
				SetState(enums.WorkflowProposalStateRejected).
				Exec(skipCtx); err != nil {
				return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionUpdate, Object: "workflowproposal"})
			}
		}
	}

	if err := closeWorkflowAssignments(ctx, r.db, instance.ID, instance.OwnerID, enums.WorkflowAssignmentStatusApproved, userID, nil); err != nil {
		return nil, err
	}

	if err := r.db.WorkflowInstance.UpdateOneID(instance.ID).
		SetState(enums.WorkflowInstanceStateCompleted).
		Exec(skipCtx); err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionUpdate, Object: "workflowinstance"})
	}

	updated, err := r.db.WorkflowInstance.Get(skipCtx, instance.ID)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionGet, Object: "workflowinstance"})
	}

	if err := recordWorkflowInstanceAdminCompletion(ctx, r.db, updated, userID, applied, nil); err != nil {
		return nil, err
	}

	return updated, nil
}

func (r *mutationResolver) cancelWorkflowInstance(ctx context.Context, id string, reason *string) (*generated.WorkflowInstance, error) {
	allowCtx := workflows.AllowContext(ctx)
	instance, err := r.db.WorkflowInstance.Get(allowCtx, id)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionGet, Object: "workflowinstance"})
	}

	if err := r.requireWorkflowAdmin(ctx, instance.OwnerID); err != nil {
		return nil, err
	}

	skipCtx := workflows.WithSkipEventEmission(allowCtx)
	workflows.MarkSkipEventEmission(skipCtx)

	if instance.OwnerID != "" {
		if err := common.SetOrganizationInAuthContext(skipCtx, &instance.OwnerID); err != nil {
			return nil, err
		}
	}

	userID, _ := auth.GetSubjectIDFromContext(ctx)

	if instance.WorkflowProposalID != "" {
		if err := r.db.WorkflowProposal.UpdateOneID(instance.WorkflowProposalID).
			SetState(enums.WorkflowProposalStateRejected).
			Exec(skipCtx); err != nil {
			return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionUpdate, Object: "workflowproposal"})
		}
	}

	if err := closeWorkflowAssignments(ctx, r.db, instance.ID, instance.OwnerID, enums.WorkflowAssignmentStatusRejected, userID, reason); err != nil {
		return nil, err
	}

	if err := r.db.WorkflowInstance.UpdateOneID(instance.ID).
		SetState(enums.WorkflowInstanceStateFailed).
		Exec(skipCtx); err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionUpdate, Object: "workflowinstance"})
	}

	updated, err := r.db.WorkflowInstance.Get(skipCtx, instance.ID)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionGet, Object: "workflowinstance"})
	}

	if err := recordWorkflowInstanceAdminCompletion(ctx, r.db, updated, userID, false, reason); err != nil {
		return nil, err
	}

	return updated, nil
}
