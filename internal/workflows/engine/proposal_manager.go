package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/workflowproposal"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/internal/workflows/observability"
	"github.com/theopenlane/iam/auth"
)

// ProposalManager handles WorkflowProposal operations
type ProposalManager struct {
	// client is the ent database client for proposal operations
	client *generated.Client
}

// NewProposalManager creates a new proposal manager
func NewProposalManager(client *generated.Client) *ProposalManager {
	return &ProposalManager{client: client}
}

// Create creates a WorkflowProposal for the approval domain within a transaction
func (m *ProposalManager) Create(ctx context.Context, tx *generated.Tx, objRef *generated.WorkflowObjectRef, domain *workflows.DomainChanges) (*generated.WorkflowProposal, error) {
	proposedHash, err := workflows.ComputeProposalHash(domain.Changes)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToComputeProposalHash, err)
	}

	now := time.Now()

	proposal, err := tx.WorkflowProposal.Create().
		SetWorkflowObjectRefID(objRef.ID).
		SetDomainKey(domain.DomainKey).
		SetState(enums.WorkflowProposalStateDraft).
		SetChanges(domain.Changes).
		SetRevision(1).
		SetProposedHash(proposedHash).
		SetUpdatedAt(now).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToCreateProposal, err)
	}

	return proposal, nil
}

// LoadForObject loads a WorkflowProposal for the given object using the ObjectFromRef registry
func (m *ProposalManager) LoadForObject(ctx context.Context, obj *workflows.Object) (*generated.WorkflowProposal, error) {
	if obj == nil || obj.ID == "" {
		return nil, nil
	}

	allowCtx, orgID, err := workflows.AllowContextWithOrg(ctx)
	if err != nil {
		return nil, err
	}

	objRefIDs, err := workflows.ObjectRefIDs(allowCtx, m.client, obj)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToQueryObjectRefs, err)
	}

	if len(objRefIDs) == 0 {
		return nil, nil
	}

	proposals, err := m.client.WorkflowProposal.Query().
		Where(
			workflowproposal.WorkflowObjectRefIDIn(objRefIDs...),
			workflowproposal.StateIn(enums.WorkflowProposalStateSubmitted, enums.WorkflowProposalStateDraft),
			workflowproposal.OwnerIDEQ(orgID),
		).
		All(allowCtx)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToQueryProposals, err)
	}

	if len(proposals) == 0 {
		return nil, nil
	}

	return proposals[0], nil
}

// ComputeHash computes a hash of the proposed changes for the given object
func (m *ProposalManager) ComputeHash(ctx context.Context, obj *workflows.Object) (string, error) {
	if obj == nil {
		return "", ErrObjectNil
	}

	proposal, err := m.LoadForObject(ctx, obj)
	if err != nil {
		return "", err
	}

	if proposal == nil {
		return "", nil
	}

	return workflows.ComputeProposalHash(proposal.Changes)
}

// Apply applies the approved changes from a WorkflowProposal to the target object
func (m *ProposalManager) Apply(scope *observability.Scope, proposalID string, obj *workflows.Object) error {
	ctx := scope.Context()

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return err
	}

	proposal, err := m.client.WorkflowProposal.Query().
		Where(
			workflowproposal.IDEQ(proposalID),
			workflowproposal.OwnerIDEQ(orgID),
		).
		Only(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToLoadProposal, err)
	}

	bypassCtx := workflows.WithContext(ctx)
	if err := workflows.ApplyObjectFieldUpdates(bypassCtx, m.client, obj.Type, obj.ID, proposal.Changes); err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToApplyFieldUpdates, err)
	}

	allowCtx := workflows.AllowContext(ctx)
	if err := m.client.WorkflowProposal.UpdateOne(proposal).
		SetState(enums.WorkflowProposalStateApplied).
		SetApprovedHash(proposal.ProposedHash).
		Exec(allowCtx); err != nil {
		scope.Warn(err, observability.Fields{
			"proposal_id": proposalID,
		})
	}

	return nil
}
