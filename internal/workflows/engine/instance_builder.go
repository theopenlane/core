package engine

import (
	"context"
	"errors"
	"fmt"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/workflowassignment"
	"github.com/theopenlane/core/internal/ent/generated/workflowinstance"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/internal/workflows/observability"
)

// createInstanceTx builds a workflow instance transaction and optional proposal
func (e *WorkflowEngine) createInstanceTx(ctx context.Context, def *generated.WorkflowDefinition, obj *workflows.Object, domain *workflows.DomainChanges, defSnapshot models.WorkflowDefinitionDocument, contextData models.WorkflowInstanceContext, ownerID string, scope *observability.Scope) (*generated.WorkflowInstance, error) {
	return workflows.WithTx(ctx, e.client, scope, func(tx *generated.Tx) (*generated.WorkflowInstance, error) {
		instance, objRef, err := workflows.CreateWorkflowInstanceWithObjectRef(ctx, tx, workflows.WorkflowInstanceBuilderParams{
			WorkflowDefinitionID: def.ID,
			DefinitionSnapshot:   defSnapshot,
			State:                enums.WorkflowInstanceStateRunning,
			Context:              contextData,
			OwnerID:              ownerID,
			ObjectType:           obj.Type,
			ObjectID:             obj.ID,
		})
		if err != nil {
			return nil, wrapCreationError(err)
		}

		if scope != nil {
			scope.WithFields(observability.Fields{
				workflowassignment.FieldWorkflowInstanceID: instance.ID,
			})
		}

		proposalID, err := e.ensureProposal(ctx, tx, def, obj, objRef, domain)
		if err != nil {
			return nil, err
		}

		if proposalID == "" {
			return instance, nil
		}

		if err := tx.WorkflowInstance.UpdateOne(instance).
			SetWorkflowProposalID(proposalID).
			Exec(ctx); err != nil {
			return nil, fmt.Errorf("failed to update instance with proposal ID: %w", err)
		}

		instance.WorkflowProposalID = proposalID
		if scope != nil {
			scope.WithFields(observability.Fields{
				workflowinstance.FieldWorkflowProposalID: proposalID,
			})
		}

		return instance, nil
	})
}

// wrapCreationError normalizes workflow instance creation errors
func wrapCreationError(err error) error {
	var creationErr *workflows.WorkflowCreationError
	if errors.As(err, &creationErr) {
		switch creationErr.Stage {
		case workflows.WorkflowCreationStageInstance:
			return fmt.Errorf("failed to create workflow instance: %w", creationErr)
		case workflows.WorkflowCreationStageObjectRef:
			return fmt.Errorf("failed to create object reference: %w", creationErr)
		}
	}

	return fmt.Errorf("failed to create workflow instance: %w", err)
}

// ensureProposal creates or reuses a workflow proposal when needed
func (e *WorkflowEngine) ensureProposal(ctx context.Context, tx *generated.Tx, def *generated.WorkflowDefinition, obj *workflows.Object, objRef *generated.WorkflowObjectRef, domain *workflows.DomainChanges) (string, error) {
	if !workflows.DefinitionHasApprovalAction(def.DefinitionJSON) || domain == nil || len(domain.Changes) == 0 {
		return "", nil
	}

	objRefIDs, err := workflows.ObjectRefIDs(ctx, e.client, obj)
	if err != nil {
		return "", fmt.Errorf("failed to query object refs: %w", err)
	}

	existingProposal, err := workflows.FindProposalForObjectRefs(
		ctx,
		e.client,
		objRefIDs,
		domain.DomainKey,
		[]enums.WorkflowProposalState{enums.WorkflowProposalStateSubmitted},
		[]enums.WorkflowProposalState{
			enums.WorkflowProposalStateSubmitted,
			enums.WorkflowProposalStateDraft,
		},
	)
	if err != nil {
		return "", err
	}

	if existingProposal != nil {
		return existingProposal.ID, nil
	}

	proposal, err := e.proposalManager.Create(ctx, tx, objRef, domain)
	if err != nil {
		return "", err
	}
	if proposal == nil {
		return "", nil
	}

	return proposal.ID, nil
}
