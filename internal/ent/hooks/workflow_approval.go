package hooks

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/workflowdefinition"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/internal/workflows"
)

// HookWorkflowApprovalRouting intercepts mutations on workflowable schemas and routes them
// to WorkflowProposal when a matching workflow definition with approval requirements exists.
// This enables the "proposed changes" pattern where mutations require approval before being applied.
func HookWorkflowApprovalRouting() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			// Skip if workflow bypass is set
			if workflows.IsWorkflowBypass(ctx) {
				return next.Mutate(ctx, m)
			}

			// Cast to GenericMutation to access common mutation methods
			mut, ok := m.(utils.GenericMutation)
			if !ok {
				return next.Mutate(ctx, m)
			}

			client := mut.Client()
			if client == nil {
				return next.Mutate(ctx, m)
			}
			if !workflowEngineEnabled(ctx, client) {
				return next.Mutate(ctx, m)
			}

			changedFields := workflows.CollectChangedFields(mut)
			if len(changedFields) == 0 {
				return next.Mutate(ctx, m)
			}

			// Check for matching workflow definition with approval requirements
			matchingDef, err := findApprovalWorkflowDefinition(ctx, client, mut, changedFields)
			if err != nil {
				return nil, err
			}
			if matchingDef == nil {
				return next.Mutate(ctx, m)
			}

			// Route to proposed changes instead of applying directly
			return routeMutationToProposal(ctx, client, mut, changedFields, matchingDef)
		})
	}, ent.OpUpdate|ent.OpUpdateOne)
}

// findApprovalWorkflowDefinition finds an active workflow definition that requires approval for the given mutation
func findApprovalWorkflowDefinition(ctx context.Context, client *generated.Client, m utils.GenericMutation, changedFields []string) (*generated.WorkflowDefinition, error) {
	// Use privacy bypass for internal workflow queries
	allowCtx, orgID, err := workflows.AllowContextWithOrg(ctx)
	if err != nil {
		return nil, err
	}

	// Query for active workflow definitions matching this schema type and operation
	query := client.WorkflowDefinition.Query().
		Where(
			workflowdefinition.SchemaTypeEQ(m.Type()),
			workflowdefinition.ActiveEQ(true),
			workflowdefinition.DraftEQ(false),
			workflowdefinition.OwnerIDEQ(orgID),
		)

	defs, err := query.All(allowCtx)
	if err != nil || len(defs) == 0 {
		return nil, err
	}

	for _, def := range defs {
		if !workflows.DefinitionHasApprovalAction(def.DefinitionJSON) {
			continue
		}

		if !workflows.DefinitionMatchesTrigger(def.DefinitionJSON, "UPDATE", changedFields) {
			continue
		}

		domains, err := workflows.ApprovalDomains(def.DefinitionJSON)
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Str("workflow_definition_id", def.ID).Msg("invalid approval action params")
			continue
		}

		if definitionTouchesChangedFields(domains, changedFields) {
			return def, nil
		}
	}

	return nil, nil
}

// definitionTouchesChangedFields checks if any of the changed fields are in the definition's approval domains
func definitionTouchesChangedFields(domains [][]string, changedFields []string) bool {
	changedSet := lo.SliceToMap(changedFields, func(f string) (string, struct{}) { return f, struct{}{} })

	return lo.SomeBy(domains, func(domain []string) bool {
		return lo.SomeBy(domain, func(field string) bool {
			_, ok := changedSet[field]
			return ok
		})
	})
}

// routeMutationToProposal stores the mutation in a WorkflowProposal instead of applying it directly
func routeMutationToProposal(ctx context.Context, client *generated.Client, m utils.GenericMutation, changedFields []string, def *generated.WorkflowDefinition) (ent.Value, error) {
	user, err := auth.GetAuthenticatedUserFromContext(ctx)
	if err != nil {
		return nil, ErrFailedToGetUserFromContext
	}

	if err := validateWorkflowEligibleFields(m.Type(), changedFields); err != nil {
		return nil, err
	}

	// Extract proposed changes from mutation
	proposedChanges := workflows.BuildProposedChanges(m, changedFields)

	objectType := enums.ToWorkflowObjectType(m.Type())
	if objectType == nil {
		return nil, ErrProposedChangesNotSupported
	}

	id, ok := getSingleMutationID(ctx, m)
	if !ok {
		return nil, ErrMutationMissingID
	}

	if len(proposedChanges) == 0 {
		allowCtx := workflows.AllowContext(ctx)
		return workflows.LoadWorkflowObject(allowCtx, client, m.Type(), id)
	}

	if err := stageWorkflowProposals(ctx, client, def, *objectType, id, proposedChanges, user.SubjectID); err != nil {
		return nil, err
	}

	// Return the existing entity (unchanged) instead of the mutated version
	allowCtx := workflows.AllowContext(ctx)
	return workflows.LoadWorkflowObject(allowCtx, client, m.Type(), id)
}

// validateWorkflowEligibleFields checks that all changed fields are eligible for workflow processing
func validateWorkflowEligibleFields(schemaType string, changedFields []string) error {
	objectType := enums.ToWorkflowObjectType(schemaType)
	if objectType == nil {
		return ErrWorkflowUnknownSchemaType
	}

	eligible := workflows.EligibleWorkflowFields(*objectType)
	if len(eligible) == 0 {
		return ErrWorkflowNoEligibleFields
	}

	for _, field := range changedFields {
		if _, ok := eligible[field]; !ok {
			return ErrWorkflowIneligibleField
		}
	}

	return nil
}

// stageProposalChanges creates or updates WorkflowProposal records for each domain
func stageProposalChanges(ctx context.Context, client *generated.Client, def *generated.WorkflowDefinition, objectType enums.WorkflowObjectType, objectID string, domainChanges []workflows.DomainChanges, userID string) error {
	// Use privacy bypass for internal workflow operations
	allowCtx := workflows.AllowContext(ctx)

	initialState := lo.Ternary(
		def.ApprovalSubmissionMode == enums.WorkflowApprovalSubmissionModeAutoSubmit,
		enums.WorkflowProposalStateSubmitted,
		enums.WorkflowProposalStateDraft,
	)

	ownerID, err := workflows.ObjectOwnerID(allowCtx, client, objectType, objectID)
	if err != nil {
		return ErrFailedToGetObjectOwnerID
	}

	obj := &workflows.Object{ID: objectID, Type: objectType}
	objRefIDs, err := workflows.ObjectRefIDs(allowCtx, client, obj)
	if err != nil {
		return ErrFailedToQueryObjectRefs
	}

	now := time.Now()

	for _, domain := range domainChanges {
		proposedHash, err := workflows.ComputeProposalHash(domain.Changes)
		if err != nil {
			return ErrFailedToComputeProposalHash
		}

		existing, err := workflows.FindProposalForObjectRefs(allowCtx, client, objRefIDs, domain.DomainKey, nil, []enums.WorkflowProposalState{
			enums.WorkflowProposalStateDraft,
			enums.WorkflowProposalStateSubmitted,
		})
		if err != nil {
			return ErrFailedToQueryExistingProposal
		}

		if existing != nil {
			if err := updateExistingProposal(ctx, def, existing, domain.Changes, proposedHash, userID, now); err != nil {
				return err
			}

			continue
		}

		newObjRefID, err := createProposalWithInstance(ctx, client, def, objectType, objectID, domain, proposedHash, ownerID, userID, initialState, now)
		if err != nil {
			return err
		}

		if newObjRefID != "" {
			objRefIDs = append(objRefIDs, newObjRefID)
		}
	}

	return nil
}

// stageWorkflowProposals stages proposals for the given proposed changes
func stageWorkflowProposals(ctx context.Context, client *generated.Client, def *generated.WorkflowDefinition, objectType enums.WorkflowObjectType, objectID string, proposedChanges map[string]any, userID string) error {
	domainChanges, err := workflows.DomainChangesForDefinition(def.DefinitionJSON, proposedChanges)
	if err != nil {
		return err
	}
	if len(domainChanges) == 0 {
		return nil
	}

	return stageProposalChanges(ctx, client, def, objectType, objectID, domainChanges, userID)
}

// updateExistingProposal updates an existing proposal with new changes
func updateExistingProposal(ctx context.Context, def *generated.WorkflowDefinition, existing *generated.WorkflowProposal, changes map[string]any, proposedHash, userID string, now time.Time) error {
	// Use privacy bypass for internal workflow operations
	allowCtx := workflows.AllowContext(ctx)

	updater := existing.Update().
		SetChanges(changes).
		SetProposedHash(proposedHash).
		SetRevision(existing.Revision + 1)

	if def.ApprovalSubmissionMode == enums.WorkflowApprovalSubmissionModeAutoSubmit || existing.State == enums.WorkflowProposalStateSubmitted {
		updater = updater.
			SetState(enums.WorkflowProposalStateSubmitted).
			SetSubmittedAt(now).
			SetSubmittedByUserID(userID)
	}

	if err := updater.Exec(allowCtx); err != nil {
		return ErrFailedToUpdateProposal
	}

	return nil
}

// createProposalWithInstance creates a new WorkflowInstance, WorkflowObjectRef, and WorkflowProposal in a transaction
func createProposalWithInstance(ctx context.Context, client *generated.Client, def *generated.WorkflowDefinition, objectType enums.WorkflowObjectType, objectID string, domain workflows.DomainChanges, proposedHash, ownerID, userID string, initialState enums.WorkflowProposalState, now time.Time) (string, error) {
	// Use privacy bypass for internal workflow operations
	allowCtx := workflows.AllowContext(ctx)

	objRefID, err := workflows.WithTx(allowCtx, client, nil, func(tx *generated.Tx) (string, error) {
		instance, objRef, createErr := workflows.CreateWorkflowInstanceWithObjectRef(allowCtx, tx, workflows.WorkflowInstanceBuilderParams{
			WorkflowDefinitionID: def.ID,
			DefinitionSnapshot:   def.DefinitionJSON,
			State:                enums.WorkflowInstanceStatePaused,
			Context: models.WorkflowInstanceContext{
				WorkflowDefinitionID: def.ID,
				ObjectType:           objectType,
				ObjectID:             objectID,
				Version:              1,
				Assignments:          []models.WorkflowAssignmentContext{},
			},
			OwnerID:    ownerID,
			ObjectType: objectType,
			ObjectID:   objectID,
		})
		if createErr != nil {
			if creationErr, ok := createErr.(*workflows.WorkflowCreationError); ok {
				switch creationErr.Stage {
				case workflows.WorkflowCreationStageInstance:
					return "", ErrFailedToCreateWorkflowInstance
				case workflows.WorkflowCreationStageObjectRef:
					return "", ErrFailedToCreateWorkflowObjectRef
				}
			}
			return "", createErr
		}

		// Create proposal
		proposalCreate := tx.WorkflowProposal.Create().
			SetWorkflowObjectRefID(objRef.ID).
			SetDomainKey(domain.DomainKey).
			SetState(initialState).
			SetChanges(domain.Changes).
			SetRevision(1).
			SetProposedHash(proposedHash).
			SetOwnerID(ownerID)

		if initialState == enums.WorkflowProposalStateSubmitted {
			proposalCreate = proposalCreate.
				SetSubmittedAt(now).
				SetSubmittedByUserID(userID)
		}

		proposal, err := proposalCreate.Save(allowCtx)
		if err != nil {
			return "", fmt.Errorf("%w: %v", ErrFailedToCreateWorkflowProposal, err)
		}

		// Link proposal to instance
		if err := tx.WorkflowInstance.UpdateOneID(instance.ID).
			SetWorkflowProposalID(proposal.ID).
			Exec(allowCtx); err != nil {
			return "", ErrFailedToLinkProposalToInstance
		}

		log.Ctx(ctx).Info().Str("proposal_id", proposal.ID).Str("object_id", objectID).Msg("proposal created")

		return objRef.ID, nil
	})
	if err != nil {
		switch {
		case errors.Is(err, workflows.ErrTxBegin):
			return "", ErrFailedToBeginTransaction
		case errors.Is(err, workflows.ErrTxCommit):
			return "", ErrFailedToCommitProposalTransaction
		}
		return "", err
	}

	return objRefID, nil
}
