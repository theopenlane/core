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
	"github.com/theopenlane/core/internal/ent/generated/workflowinstance"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/internal/ent/workflowgenerated"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/internal/workflows/engine"
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

			allChangedFields := workflows.CollectAllChangedFields(mut)
			changedFields := workflows.CollectChangedFields(mut)
			changedEdges, addedIDs, removedIDs := workflowgenerated.ExtractChangedEdges(m)
			if len(allChangedFields) == 0 && len(changedEdges) == 0 {
				return next.Mutate(ctx, m)
			}

			wfEngine, _ := client.WorkflowEngine.(*engine.WorkflowEngine)
			if wfEngine == nil {
				if ctxClient := generated.FromContext(ctx); ctxClient != nil {
					wfEngine, _ = ctxClient.WorkflowEngine.(*engine.WorkflowEngine)
				}
			}
			if wfEngine == nil {
				return next.Mutate(ctx, m)
			}

			objectType := enums.ToWorkflowObjectType(mut.Type())
			if objectType == nil {
				return next.Mutate(ctx, m)
			}

			id, ok := getSingleMutationID(ctx, mut)
			if !ok {
				return next.Mutate(ctx, m)
			}

			allowCtx := workflows.AllowContext(ctx)
			entity, err := workflows.LoadWorkflowObject(allowCtx, client, mut.Type(), id)
			if err != nil {
				return nil, err
			}

			obj := &workflows.Object{
				ID:   id,
				Type: *objectType,
				Node: entity,
			}

			proposedChanges := workflows.BuildProposedChanges(mut, changedFields)
			if len(proposedChanges) == 0 {
				return next.Mutate(ctx, m)
			}

			definitions, err := wfEngine.FindMatchingDefinitions(allowCtx, mut.Type(), "UPDATE", changedFields, changedEdges, addedIDs, removedIDs, proposedChanges, obj)
			if err != nil || len(definitions) == 0 {
				return next.Mutate(ctx, m)
			}

			preCommitDefs := make([]*generated.WorkflowDefinition, 0, len(definitions))
			for _, def := range definitions {
				if !workflows.DefinitionHasApprovalAction(def.DefinitionJSON) {
					continue
				}

				shouldRun, err := wfEngine.EvaluateConditions(allowCtx, def, obj, "UPDATE", changedFields, changedEdges, addedIDs, removedIDs, proposedChanges)
				if err != nil {
					return nil, err
				}
				if !shouldRun {
					continue
				}

				if !workflows.DefinitionUsesPreCommitApprovals(def.DefinitionJSON) {
					continue
				}

				preCommitDefs = append(preCommitDefs, def)
			}

			if len(preCommitDefs) == 0 {
				return next.Mutate(ctx, m)
			}

			eligibleFields, ineligibleFields := workflows.SeparateFieldsByEligibility(mut.Type(), allChangedFields)
			ineligibleFields = filterNonSystemFields(ineligibleFields)

			eligibleChanges := workflows.BuildProposedChanges(mut, eligibleFields)
			hasDirectChanges := len(ineligibleFields) > 0 || len(changedEdges) > 0
			if hasDirectChanges {
				resetMutationFields(m, eligibleFields)
				bypassCtx := workflows.WithContext(ctx)
				if _, err := next.Mutate(bypassCtx, m); err != nil {
					return nil, err
				}
			}

			if len(eligibleFields) == 0 {
				return workflows.LoadWorkflowObject(allowCtx, client, mut.Type(), id)
			}

			proposedChanges = eligibleChanges

			// Route to proposed changes instead of applying directly
			workflows.MarkSkipEventEmission(ctx)
			return routeMutationToProposals(ctx, client, mut, proposedChanges, preCommitDefs)
		})
	}, ent.OpUpdate|ent.OpUpdateOne)
}

// routeMutationToProposals stores the mutation in WorkflowProposal records instead of applying it directly.
func routeMutationToProposals(ctx context.Context, client *generated.Client, m utils.GenericMutation, proposedChanges map[string]any, defs []*generated.WorkflowDefinition) (ent.Value, error) {
	user, err := auth.GetAuthenticatedUserFromContext(ctx)
	if err != nil {
		return nil, ErrFailedToGetUserFromContext
	}

	objectType := enums.ToWorkflowObjectType(m.Type())
	if objectType == nil {
		return nil, ErrProposedChangesNotSupported
	}

	id, ok := getSingleMutationID(ctx, m)
	if !ok {
		return nil, ErrMutationMissingID
	}

	allowCtx := workflows.AllowContext(ctx)

	if len(proposedChanges) == 0 {
		return workflows.LoadWorkflowObject(allowCtx, client, m.Type(), id)
	}

	// Load the existing entity BEFORE staging proposals and triggering workflows.
	// This ensures we return the original (unchanged) entity even if workflows
	// auto-apply proposals synchronously.
	originalEntity, err := workflows.LoadWorkflowObject(allowCtx, client, m.Type(), id)
	if err != nil {
		return nil, err
	}

	for _, def := range defs {
		if err := stageWorkflowProposals(ctx, client, def, *objectType, id, proposedChanges, user.SubjectID); err != nil {
			return nil, err
		}
	}

	return originalEntity, nil
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
		if _, ok := workflowApprovalIgnoredFields[field]; ok {
			continue
		}
		if _, ok := eligible[field]; !ok {
			return fmt.Errorf("%w: %s", ErrWorkflowIneligibleField, field)
		}
	}

	return nil
}

var workflowApprovalIgnoredFields = map[string]struct{}{
	"created_at": {},
	"created_by": {},
	"revision":   {},
	"summary":    {},
	"display_id": {},
	"updated_at": {},
	"updated_by": {},
	"deleted_at": {},
	"deleted_by": {},
}

// filterNonSystemFields removes system fields from the given field list
func filterNonSystemFields(fields []string) []string {
	if len(fields) == 0 {
		return fields
	}
	remaining := make([]string, 0, len(fields))
	for _, field := range fields {
		if _, ok := workflowApprovalIgnoredFields[field]; ok {
			continue
		}
		remaining = append(remaining, field)
	}
	return remaining
}

// resetMutationFields removes any pending changes for the provided field names.
// It relies on ent-generated Reset<Field> methods, falling back silently when no reset exists.
func resetMutationFields(m ent.Mutation, fields []string) {
	if m == nil || len(fields) == 0 {
		return
	}

	for _, field := range fields {
		// Ignore unknown fields to preserve the prior "silent fallback" behavior.
		_ = m.ResetField(field)
	}
}

// resolveApprovalSubmissionMode returns the effective approval submission mode.
// Defaults to AUTO_SUBMIT when not explicitly set in the definition JSON.
func resolveApprovalSubmissionMode(def *generated.WorkflowDefinition) enums.WorkflowApprovalSubmissionMode {
	if def == nil || def.DefinitionJSON.ApprovalSubmissionMode == "" {
		if def == nil {
			return enums.WorkflowApprovalSubmissionModeAutoSubmit
		}

		// Fall back to the persisted definition column when the JSON omits it.
		if def.ApprovalSubmissionMode != "" {
			if parsed := enums.ToWorkflowApprovalSubmissionMode(def.ApprovalSubmissionMode.String()); parsed != nil {
				if *parsed == enums.WorkflowApprovalSubmissionModeManualSubmit {
					return enums.WorkflowApprovalSubmissionModeAutoSubmit
				}
				return *parsed
			}
		}

		return enums.WorkflowApprovalSubmissionModeAutoSubmit
	}

	if parsed := enums.ToWorkflowApprovalSubmissionMode(def.DefinitionJSON.ApprovalSubmissionMode.String()); parsed != nil {
		if *parsed == enums.WorkflowApprovalSubmissionModeManualSubmit {
			return enums.WorkflowApprovalSubmissionModeAutoSubmit
		}
		return *parsed
	}

	return enums.WorkflowApprovalSubmissionModeAutoSubmit
}

// stageProposalChanges creates or updates WorkflowProposal records for each domain
func stageProposalChanges(ctx context.Context, client *generated.Client, def *generated.WorkflowDefinition, objectType enums.WorkflowObjectType, objectID string, domainChanges []workflows.DomainChanges, userID string) error {
	// Use privacy bypass for internal workflow operations
	allowCtx := workflows.AllowContext(ctx)

	submissionMode := resolveApprovalSubmissionMode(def)
	initialState := lo.Ternary(
		submissionMode == enums.WorkflowApprovalSubmissionModeAutoSubmit,
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

			instance, err := ensureInstanceForExistingProposal(ctx, client, def, objectType, objectID, ownerID, existing)
			if err != nil {
				return err
			}

			if submissionMode == enums.WorkflowApprovalSubmissionModeAutoSubmit && instance != nil {
				if err := triggerWorkflowAfterProposalCreation(ctx, client, def, objectType, objectID, domain, instance.ID); err != nil {
					log.Ctx(ctx).Error().Err(err).Str("instance_id", instance.ID).Msg("failed to trigger workflow after proposal update")
				}
			}

			continue
		}

		result, err := createProposalWithInstance(ctx, client, def, objectType, objectID, domain, proposedHash, ownerID, userID, initialState, now)
		if err != nil {
			return err
		}

		if result != nil && result.ObjRefID != "" {
			objRefIDs = append(objRefIDs, result.ObjRefID)
		}
	}

	return nil
}

// ensureInstanceForExistingProposal makes sure a workflow instance exists for the definition + proposal.
func ensureInstanceForExistingProposal(ctx context.Context, client *generated.Client, def *generated.WorkflowDefinition, objectType enums.WorkflowObjectType, objectID, ownerID string, proposal *generated.WorkflowProposal) (*generated.WorkflowInstance, error) {
	// Use privacy bypass for internal workflow operations
	allowCtx := workflows.AllowContext(ctx)

	instance, err := client.WorkflowInstance.Query().
		Where(
			workflowinstance.WorkflowDefinitionIDEQ(def.ID),
			workflowinstance.WorkflowProposalIDEQ(proposal.ID),
		).
		First(allowCtx)
	if err == nil {
		return instance, nil
	}
	if !generated.IsNotFound(err) {
		return nil, err
	}

	var created *generated.WorkflowInstance
	_, err = workflows.WithTx(allowCtx, client, nil, func(tx *generated.Tx) (string, error) {
		createdInstance, _, createErr := workflows.CreateWorkflowInstanceWithObjectRef(allowCtx, tx, workflows.WorkflowInstanceBuilderParams{
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

		if err := tx.WorkflowInstance.UpdateOneID(createdInstance.ID).
			SetWorkflowProposalID(proposal.ID).
			Exec(allowCtx); err != nil {
			return "", ErrFailedToLinkProposalToInstance
		}

		created = createdInstance
		return "", nil
	})
	if err != nil {
		switch {
		case errors.Is(err, workflows.ErrTxBegin):
			return nil, ErrFailedToBeginTransaction
		case errors.Is(err, workflows.ErrTxCommit):
			return nil, ErrFailedToCommitProposalTransaction
		}
		return nil, err
	}

	return created, nil
}

// stageWorkflowProposals stages proposals for the given proposed changes
func stageWorkflowProposals(ctx context.Context, client *generated.Client, def *generated.WorkflowDefinition, objectType enums.WorkflowObjectType, objectID string, proposedChanges map[string]any, userID string) error {
	domainChanges, err := workflows.DomainChangesForDefinition(def.DefinitionJSON, objectType, proposedChanges)
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

	submissionMode := resolveApprovalSubmissionMode(def)
	updater := existing.Update().
		SetChanges(changes).
		SetProposedHash(proposedHash).
		SetRevision(existing.Revision + 1)

	if submissionMode == enums.WorkflowApprovalSubmissionModeAutoSubmit || existing.State == enums.WorkflowProposalStateSubmitted {
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

// proposalCreationResult holds the results of creating a proposal with instance
type proposalCreationResult struct {
	ObjRefID   string
	InstanceID string
	ProposalID string
}

// createProposalWithInstance creates a new WorkflowInstance, WorkflowObjectRef, and WorkflowProposal in a transaction
func createProposalWithInstance(ctx context.Context, client *generated.Client, def *generated.WorkflowDefinition, objectType enums.WorkflowObjectType, objectID string, domain workflows.DomainChanges, proposedHash, ownerID, userID string, initialState enums.WorkflowProposalState, now time.Time) (*proposalCreationResult, error) {
	// Use privacy bypass for internal workflow operations
	allowCtx := workflows.AllowContext(ctx)

	var result proposalCreationResult

	_, err := workflows.WithTx(allowCtx, client, nil, func(tx *generated.Tx) (string, error) {
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

		// Create proposal as DRAFT first so the submit hook can see the instance link
		// before triggering/resuming workflows. If auto-submitting, we update to
		// SUBMITTED after the link is established.
		proposal, err := tx.WorkflowProposal.Create().
			SetWorkflowObjectRefID(objRef.ID).
			SetDomainKey(domain.DomainKey).
			SetState(enums.WorkflowProposalStateDraft).
			SetChanges(domain.Changes).
			SetRevision(1).
			SetProposedHash(proposedHash).
			SetOwnerID(ownerID).
			Save(allowCtx)
		if err != nil {
			return "", fmt.Errorf("%w: %v", ErrFailedToCreateWorkflowProposal, err)
		}

		// Link proposal to instance BEFORE updating state to SUBMITTED
		if err := tx.WorkflowInstance.UpdateOneID(instance.ID).
			SetWorkflowProposalID(proposal.ID).
			Exec(allowCtx); err != nil {
			return "", ErrFailedToLinkProposalToInstance
		}

		// Now update to SUBMITTED if that's the initial state. Use bypass context to
		// skip HookWorkflowProposalTriggerOnSubmit since we trigger the workflow
		// after the transaction commits.
		if initialState == enums.WorkflowProposalStateSubmitted {
			bypassCtx := workflows.AllowBypassContext(allowCtx)
			if err := tx.WorkflowProposal.UpdateOneID(proposal.ID).
				SetState(enums.WorkflowProposalStateSubmitted).
				SetSubmittedAt(now).
				SetSubmittedByUserID(userID).
				Exec(bypassCtx); err != nil {
				return "", ErrFailedToUpdateWorkflowProposal
			}
		}

		log.Ctx(ctx).Info().Str("proposal_id", proposal.ID).Str("object_id", objectID).Msg("proposal created")

		result = proposalCreationResult{
			ObjRefID:   objRef.ID,
			InstanceID: instance.ID,
			ProposalID: proposal.ID,
		}

		return objRef.ID, nil
	})
	if err != nil {
		switch {
		case errors.Is(err, workflows.ErrTxBegin):
			return nil, ErrFailedToBeginTransaction
		case errors.Is(err, workflows.ErrTxCommit):
			return nil, ErrFailedToCommitProposalTransaction
		}
		return nil, err
	}

	// After transaction commits, trigger the workflow if auto-submitted
	if initialState == enums.WorkflowProposalStateSubmitted {
		if err := triggerWorkflowAfterProposalCreation(ctx, client, def, objectType, objectID, domain, result.InstanceID); err != nil {
			log.Ctx(ctx).Error().Err(err).Str("instance_id", result.InstanceID).Msg("failed to trigger workflow after proposal creation")
		}
	}

	return &result, nil
}

// triggerWorkflowAfterProposalCreation triggers a workflow for a newly created proposal
func triggerWorkflowAfterProposalCreation(ctx context.Context, client *generated.Client, def *generated.WorkflowDefinition, objectType enums.WorkflowObjectType, objectID string, domain workflows.DomainChanges, instanceID string) error {
	// Get the workflow engine from the client or context
	wfEngine, _ := client.WorkflowEngine.(*engine.WorkflowEngine)
	if wfEngine == nil {
		if ctxClient := generated.FromContext(ctx); ctxClient != nil {
			wfEngine, _ = ctxClient.WorkflowEngine.(*engine.WorkflowEngine)
		}
	}
	if wfEngine == nil {
		log.Ctx(ctx).Debug().Str("instance_id", instanceID).Msg("workflow engine not available, skipping trigger")
		return nil
	}

	allowCtx := workflows.AllowContext(ctx)

	instance, err := client.WorkflowInstance.Get(allowCtx, instanceID)
	if err != nil {
		return fmt.Errorf("failed to load instance: %w", err)
	}
	if instance.State != enums.WorkflowInstanceStatePaused {
		log.Ctx(ctx).Debug().Str("instance_id", instance.ID).Str("state", instance.State.String()).Msg("workflow instance not paused, skipping trigger")
		return nil
	}

	obj := &workflows.Object{
		ID:   objectID,
		Type: objectType,
	}

	changedFields := make([]string, 0, len(domain.Changes))
	for field := range domain.Changes {
		changedFields = append(changedFields, field)
	}

	return wfEngine.TriggerExistingInstance(allowCtx, instance, def, obj, engine.TriggerInput{
		EventType:       "UPDATE",
		ChangedFields:   changedFields,
		ProposedChanges: domain.Changes,
	})
}
