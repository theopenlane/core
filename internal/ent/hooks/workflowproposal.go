package hooks

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"time"

	"entgo.io/ent"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/workflowassignment"
	"github.com/theopenlane/core/internal/ent/generated/workflowinstance"
	"github.com/theopenlane/core/internal/ent/generated/workflowproposal"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/internal/workflows/engine"
	"github.com/theopenlane/iam/auth"
)

// HookWorkflowProposalInvalidateAssignments invalidates approved assignments when a SUBMITTED proposal is edited
func HookWorkflowProposalInvalidateAssignments() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.WorkflowProposalFunc(func(ctx context.Context, m *generated.WorkflowProposalMutation) (generated.Value, error) {
			client := m.Client()
			if !workflowEngineEnabled(ctx, client) {
				return next.Mutate(ctx, m)
			}

			id, ok := getSingleMutationID(ctx, m)
			if !ok {
				return next.Mutate(ctx, m)
			}

			// Use privacy bypass for internal workflow queries
			allowCtx, orgID, err := workflows.AllowContextWithOrg(ctx)
			if err != nil {
				return nil, err
			}

			proposal, err := client.WorkflowProposal.Query().
				Where(
					workflowproposal.ID(id),
					workflowproposal.OwnerIDEQ(orgID),
				).
				Only(allowCtx)
			if err != nil {
				log.Ctx(ctx).Debug().Err(err).Str("proposal_id", id).Msg("invalidate hook: failed to query proposal")

				return nil, ErrFailedToQueryWorkflowProposal
			}
			if proposal.State != enums.WorkflowProposalStateSubmitted {
				log.Ctx(ctx).Debug().Str("proposal_id", id).Str("state", proposal.State.String()).Msg("invalidate hook: proposal not SUBMITTED")

				return next.Mutate(ctx, m)
			}

			if !slices.Contains(m.Fields(), workflowproposal.FieldChanges) {
				log.Ctx(ctx).Debug().Str("proposal_id", id).Strs("fields", m.Fields()).Msg("invalidate hook: changes field not in mutation")

				return next.Mutate(ctx, m)
			}

			instances, err := client.WorkflowInstance.Query().
				Where(
					workflowinstance.WorkflowProposalID(id),
					workflowinstance.OwnerIDEQ(orgID),
				).
				All(allowCtx)
			if err != nil {
				return nil, ErrFailedToQueryWorkflowInstances
			}
			if len(instances) == 0 {
				return next.Mutate(ctx, m)
			}

			// Get the user making this change - may be empty for system operations
			invalidatedByUserID, err := auth.GetSubjectIDFromContext(ctx)
			if err != nil {
				log.Ctx(ctx).Debug().Err(err).Msg("invalidate hook: no user in context, using empty user ID")
			}

			// Get the old hash before mutation
			oldProposedHash := proposal.ProposedHash

			// Compute new hash from the incoming changes
			var newProposedHash string
			if newChanges, ok := m.Changes(); ok {
				newProposedHash, err = workflows.ComputeProposalHash(newChanges)
				if err != nil {
					return nil, err
				}
			}

			for _, instance := range instances {
				if err := invalidateInstanceAssignments(allowCtx, client, instance, proposal, invalidatedByUserID, oldProposedHash, newProposedHash); err != nil {
					return nil, err
				}
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpUpdate|ent.OpUpdateOne)
}

// invalidateInstanceAssignments invalidates approved assignments for a workflow instance and notifies affected users
func invalidateInstanceAssignments(ctx context.Context, client *generated.Client, instance *generated.WorkflowInstance, proposal *generated.WorkflowProposal, invalidatedByUserID, oldProposedHash, newProposedHash string) error {
	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return err
	}

	assignments, err := client.WorkflowAssignment.Query().
		Where(
			workflowassignment.WorkflowInstanceIDEQ(instance.ID),
			workflowassignment.StatusEQ(enums.WorkflowAssignmentStatusApproved),
			workflowassignment.OwnerIDEQ(orgID),
		).
		WithWorkflowAssignmentTargets().
		All(ctx)
	if err != nil {
		return ErrFailedToQueryAssignments
	}

	if len(assignments) == 0 {
		return nil
	}

	invalidatedAt := time.Now().UTC()
	var affectedUserIDs []string

	for _, assignment := range assignments {
		invalidation := models.WorkflowAssignmentInvalidation{
			Reason:              "proposal changes edited after approval",
			PreviousStatus:      enums.WorkflowAssignmentStatusApproved.String(),
			InvalidatedAt:       invalidatedAt.Format(time.RFC3339),
			InvalidatedByUserID: invalidatedByUserID,
			ApprovedHash:        oldProposedHash,
			NewProposedHash:     newProposedHash,
		}

		if err := client.WorkflowAssignment.UpdateOneID(assignment.ID).
			SetStatus(enums.WorkflowAssignmentStatusPending).
			SetInvalidationMetadata(invalidation).
			Exec(ctx); err != nil {
			return ErrFailedToInvalidateAssignment
		}

		// Collect affected user IDs from assignment targets
		for _, target := range assignment.Edges.WorkflowAssignmentTargets {
			if target.TargetUserID != "" {
				affectedUserIDs = append(affectedUserIDs, target.TargetUserID)
			}
		}
	}

	log.Ctx(ctx).Info().Str("instance_id", instance.ID).Str("proposal_id", proposal.ID).Int("count", len(assignments)).Str("invalidated_by", invalidatedByUserID).Msg("invalidated approved assignments due to proposal edit")

	if err := recordAssignmentsInvalidated(ctx, client, instance, proposal, len(assignments)); err != nil {
		return err
	}

	// Send notifications to affected users
	affectedUserIDs = lo.Uniq(affectedUserIDs)
	if err := sendInvalidationNotifications(ctx, client, instance, proposal, affectedUserIDs, invalidatedByUserID); err != nil {
		return err
	}

	return nil
}

// sendInvalidationNotifications sends notifications to users whose approvals were invalidated
func sendInvalidationNotifications(ctx context.Context, client *generated.Client, instance *generated.WorkflowInstance, proposal *generated.WorkflowProposal, userIDs []string, invalidatedByUserID string) error {
	if len(userIDs) == 0 {
		return nil
	}

	ownerID, err := workflows.ResolveOwnerID(ctx, instance.OwnerID)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Str("instance_id", instance.ID).Msg("failed to resolve owner for invalidation notification")
		return ErrFailedToResolveInvalidationNotificationOwner
	}

	title := "Approval Required: Changes Updated"
	body := "The changes you previously approved have been modified. Please review and re-approve."

	for _, userID := range userIDs {
		notificationData := map[string]any{
			"workflow_instance_id": instance.ID,
			"proposal_id":          proposal.ID,
			"domain_key":           proposal.DomainKey,
			"invalidated_by":       invalidatedByUserID,
		}

		builder := client.Notification.Create().
			SetNotificationType(enums.NotificationTypeUser).
			SetObjectType("workflow.approval_invalidated").
			SetTitle(title).
			SetBody(body).
			SetData(notificationData).
			SetUserID(userID).
			SetOwnerID(ownerID)

		if err := builder.Exec(ctx); err != nil {
			log.Ctx(ctx).Warn().Err(err).Str("user_id", userID).Str("instance_id", instance.ID).Msg("failed to send invalidation notification")
			return ErrFailedToSendInvalidationNotification
		}
	}

	return nil
}

// HookWorkflowProposalTriggerOnSubmit triggers workflows when a proposal transitions to SUBMITTED state
func HookWorkflowProposalTriggerOnSubmit() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.WorkflowProposalFunc(func(ctx context.Context, m *generated.WorkflowProposalMutation) (generated.Value, error) {
			value, err := next.Mutate(ctx, m)
			if err != nil {
				return value, err
			}

			client := m.Client()
			if !workflowEngineEnabled(ctx, client) {
				return value, nil
			}

			newState, stateUpdated := m.State()
			if !stateUpdated || newState != enums.WorkflowProposalStateSubmitted {
				return value, nil
			}

			ids, err := GetObjectIDsFromMutation(ctx, m, value)
			if err != nil {
				return value, err
			}
			var id string
			if len(ids) == 1 {
				id = ids[0]
			} else if proposal, ok := value.(*generated.WorkflowProposal); ok {
				id = proposal.ID
			}
			if id == "" {
				return value, nil
			}

			orgID, err := auth.GetOrganizationIDFromContext(ctx)
			if err != nil {
				return value, err
			}

			proposal, err := client.WorkflowProposal.Query().
				Where(
					workflowproposal.ID(id),
					workflowproposal.OwnerIDEQ(orgID),
				).
				WithWorkflowObjectRef().
				WithWorkflowInstances().
				Only(ctx)
			if err != nil {
				log.Ctx(ctx).Error().Err(err).Str("proposal_id", id).Msg("failed to load proposal for workflow trigger")
				return value, ErrFailedToLoadWorkflowProposalForTrigger
			}

			if err := triggerWorkflowForProposal(ctx, client, proposal); err != nil {
				log.Ctx(ctx).Error().Err(err).Str("proposal_id", proposal.ID).Msg("failed to trigger workflow for submitted proposal")
				return value, err
			}

			return value, nil
		})
	}, ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne)
}

// triggerWorkflowForProposal starts workflows for a submitted proposal
func triggerWorkflowForProposal(ctx context.Context, client *generated.Client, proposal *generated.WorkflowProposal) error {
	if proposal.State != enums.WorkflowProposalStateSubmitted {
		return nil
	}

	// WorkflowEngine is injected on the ent client when workflows are enabled.
	wfEngine, ok := client.WorkflowEngine.(*engine.WorkflowEngine)
	if !ok || wfEngine == nil {
		return nil
	}

	objRef := proposal.Edges.WorkflowObjectRef
	if objRef == nil {
		return ErrWorkflowProposalMissingObjectRef
	}

	obj, err := workflows.ObjectFromRef(objRef)
	if err != nil {
		return ErrFailedToDeriveObjectFromRef
	}

	changedFields := lo.Keys(proposal.Changes)

	entity, err := workflows.LoadWorkflowObject(ctx, client, obj.Type.String(), obj.ID)
	if err != nil {
		return ErrFailedToLoadWorkflowObject
	}
	obj.Node = entity

	allowCtx := workflows.AllowContext(ctx)
	definitions, err := wfEngine.FindMatchingDefinitions(allowCtx, obj.Type.String(), "UPDATE", changedFields, nil, nil, nil, proposal.Changes, obj)
	if err != nil {
		return ErrFailedToFindMatchingDefinitions
	}

	for _, def := range definitions {
		if !workflows.DefinitionHasApprovalAction(def.DefinitionJSON) {
			continue
		}

		if existing := findInstanceForDefinition(proposal.Edges.WorkflowInstances, def.ID); existing != nil {
			if err := wfEngine.TriggerExistingInstance(ctx, existing, def, obj, engine.TriggerInput{
				EventType:     "UPDATE",
				ChangedFields: changedFields,
			}); err != nil {
				log.Ctx(ctx).Error().Err(err).Str("definition_id", def.ID).Str("instance_id", existing.ID).Msg("failed to resume workflow")
				return ErrFailedToResumeWorkflowInstance
			}

			continue
		}

		_, err := wfEngine.TriggerWorkflow(ctx, def, obj, engine.TriggerInput{
			EventType:       "UPDATE",
			ChangedFields:   changedFields,
			ProposedChanges: proposal.Changes,
		})
		if err != nil && !errors.Is(err, workflows.ErrWorkflowAlreadyActive) {
			log.Ctx(ctx).Error().Err(err).Str("definition_id", def.ID).Msg("failed to trigger workflow")
			return ErrFailedToTriggerWorkflow
		}
	}

	return nil
}

// findInstanceForDefinition returns the active instance matching a workflow definition
func findInstanceForDefinition(instances []*generated.WorkflowInstance, definitionID string) *generated.WorkflowInstance {
	return lo.FindOrElse(instances, nil, func(i *generated.WorkflowInstance) bool {
		return i.WorkflowDefinitionID == definitionID &&
			i.State != enums.WorkflowInstanceStateCompleted &&
			i.State != enums.WorkflowInstanceStateFailed
	})
}

// recordAssignmentsInvalidated records invalidation events for assignments
func recordAssignmentsInvalidated(ctx context.Context, client *generated.Client, instance *generated.WorkflowInstance, proposal *generated.WorkflowProposal, invalidatedCount int) error {
	details, _ := json.Marshal(map[string]any{
		"proposal_id":       proposal.ID,
		"domain_key":        proposal.DomainKey,
		"invalidation_note": "proposal changes edited after approval",
		"invalidated_count": invalidatedCount,
	})

	create := client.WorkflowEvent.Create().
		SetWorkflowInstanceID(instance.ID).
		SetEventType(enums.WorkflowEventTypeAssignmentInvalidated).
		SetPayload(models.WorkflowEventPayload{
			EventType: enums.WorkflowEventTypeAssignmentInvalidated,
			Details:   details,
		}).
		SetOwnerID(lo.CoalesceOrEmpty(instance.OwnerID, proposal.OwnerID))

	if err := create.Exec(ctx); err != nil {
		log.Ctx(ctx).Warn().Err(err).Str("instance_id", instance.ID).Msg("failed to record assignment invalidation event")
		return ErrFailedToRecordAssignmentInvalidationEvent
	}

	return nil
}
