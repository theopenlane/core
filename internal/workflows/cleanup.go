package workflows

import (
	"context"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/workflowassignment"
	"github.com/theopenlane/core/internal/ent/generated/workflowassignmenttarget"
	"github.com/theopenlane/core/internal/ent/generated/workflowdefinition"
	"github.com/theopenlane/core/internal/ent/generated/workflowevent"
	"github.com/theopenlane/core/internal/ent/generated/workflowinstance"
	"github.com/theopenlane/core/internal/ent/generated/workflowobjectref"
	"github.com/theopenlane/core/internal/ent/generated/workflowproposal"
)

// FindOrphanWorkflowInstanceIDs returns workflow instance IDs whose definitions are missing or soft-deleted.
// If ownerID is provided, results are scoped to that organization.
func FindOrphanWorkflowInstanceIDs(ctx context.Context, client *generated.Client, ownerID string) ([]string, error) {
	if client == nil {
		return nil, ErrNilClient
	}

	allowCtx := AllowContext(ctx)

	query := client.WorkflowInstance.Query().
		Where(
			workflowinstance.Not(
				workflowinstance.HasWorkflowDefinitionWith(workflowdefinition.DeletedAtIsNil()),
			),
		)
	if ownerID != "" {
		query = query.Where(workflowinstance.OwnerIDEQ(ownerID))
	}

	return query.Select(workflowinstance.FieldID).Strings(allowCtx)
}

// DeleteWorkflowInstanceChildren removes workflow instance children such as assignments, targets, proposals, object refs, and events.
func DeleteWorkflowInstanceChildren(ctx context.Context, client *generated.Client, instanceIDs []string) error {
	if client == nil {
		return ErrNilClient
	}
	if len(instanceIDs) == 0 {
		return nil
	}

	allowCtx := AllowContext(ctx)

	assignmentIDs, err := client.WorkflowAssignment.Query().
		Where(workflowassignment.WorkflowInstanceIDIn(instanceIDs...)).
		Select(workflowassignment.FieldID).
		Strings(allowCtx)
	if err != nil {
		return err
	}

	if len(assignmentIDs) > 0 {
		if _, err := client.WorkflowAssignmentTarget.Delete().
			Where(workflowassignmenttarget.WorkflowAssignmentIDIn(assignmentIDs...)).
			Exec(allowCtx); err != nil {
			return err
		}
	}

	if _, err := client.WorkflowAssignment.Delete().
		Where(workflowassignment.WorkflowInstanceIDIn(instanceIDs...)).
		Exec(allowCtx); err != nil {
		return err
	}

	objRefIDs, err := client.WorkflowObjectRef.Query().
		Where(workflowobjectref.WorkflowInstanceIDIn(instanceIDs...)).
		Select(workflowobjectref.FieldID).
		Strings(allowCtx)
	if err != nil {
		return err
	}

	proposalIDSet := make(map[string]struct{})
	if len(objRefIDs) > 0 {
		proposalIDs, err := client.WorkflowProposal.Query().
			Where(workflowproposal.WorkflowObjectRefIDIn(objRefIDs...)).
			Select(workflowproposal.FieldID).
			Strings(allowCtx)
		if err != nil {
			return err
		}

		for _, proposalID := range proposalIDs {
			if proposalID != "" {
				proposalIDSet[proposalID] = struct{}{}
			}
		}
	}

	proposalIDsFromInstances, err := client.WorkflowInstance.Query().
		Where(
			workflowinstance.IDIn(instanceIDs...),
			workflowinstance.WorkflowProposalIDNotNil(),
		).
		Select(workflowinstance.FieldWorkflowProposalID).
		Strings(allowCtx)
	if err != nil {
		return err
	}
	for _, proposalID := range proposalIDsFromInstances {
		if proposalID != "" {
			proposalIDSet[proposalID] = struct{}{}
		}
	}

	if len(proposalIDSet) > 0 {
		proposalIDs := make([]string, 0, len(proposalIDSet))
		for proposalID := range proposalIDSet {
			proposalIDs = append(proposalIDs, proposalID)
		}
		if _, err := client.WorkflowProposal.Delete().
			Where(workflowproposal.IDIn(proposalIDs...)).
			Exec(allowCtx); err != nil {
			return err
		}
	}

	if len(objRefIDs) > 0 {
		if _, err := client.WorkflowObjectRef.Delete().
			Where(workflowobjectref.IDIn(objRefIDs...)).
			Exec(allowCtx); err != nil {
			return err
		}
	}

	if _, err := client.WorkflowEvent.Delete().
		Where(workflowevent.WorkflowInstanceIDIn(instanceIDs...)).
		Exec(allowCtx); err != nil {
		return err
	}

	return nil
}

// DeleteWorkflowInstancesCascade removes workflow instances and associated children via hooks.
func DeleteWorkflowInstancesCascade(ctx context.Context, client *generated.Client, instanceIDs []string) error {
	if client == nil {
		return ErrNilClient
	}
	if len(instanceIDs) == 0 {
		return nil
	}

	allowCtx := AllowContext(ctx)

	if _, err := client.WorkflowInstance.Delete().
		Where(workflowinstance.IDIn(instanceIDs...)).
		Exec(allowCtx); err != nil {
		return err
	}

	return nil
}

// CleanupOrphanWorkflowInstances deletes orphaned workflow instances and returns deleted IDs.
// If ownerID is provided, deletions are scoped to that organization.
func CleanupOrphanWorkflowInstances(ctx context.Context, client *generated.Client, ownerID string) ([]string, error) {
	ids, err := FindOrphanWorkflowInstanceIDs(ctx, client, ownerID)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}

	if err := DeleteWorkflowInstancesCascade(ctx, client, ids); err != nil {
		return nil, err
	}

	return ids, nil
}
