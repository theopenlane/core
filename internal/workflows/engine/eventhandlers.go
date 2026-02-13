package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/events"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/workflowassignment"
	"github.com/theopenlane/core/internal/ent/generated/workflowevent"
	"github.com/theopenlane/core/internal/ent/generated/workflowinstance"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/internal/workflows/observability"
	"github.com/theopenlane/core/pkg/events/soiree"
	"github.com/theopenlane/iam/auth"
)

// WorkflowListeners contains all workflow event listeners
type WorkflowListeners struct {
	client   *generated.Client
	engine   *WorkflowEngine
	emitter  soiree.Emitter
	observer *observability.Observer
}

// NewWorkflowListeners creates workflow event listeners.
func NewWorkflowListeners(client *generated.Client, engine *WorkflowEngine, emitter soiree.Emitter) *WorkflowListeners {
	return &WorkflowListeners{
		client:   client,
		engine:   engine,
		emitter:  emitter,
		observer: engine.observer,
	}
}

// HandleWorkflowMutation triggers matching workflows for workflow-eligible mutations.
func (l *WorkflowListeners) HandleWorkflowMutation(ctx *soiree.EventContext, payload *events.MutationPayload) error {
	if payload == nil {
		return nil
	}

	if workflows.IsWorkflowBypass(ctx.Context()) && !workflows.AllowWorkflowEventEmission(ctx.Context()) {
		return nil
	}
	if workflows.ShouldSkipEventEmission(ctx.Context()) {
		return nil
	}

	if payload.Operation != ent.OpUpdate.String() && payload.Operation != ent.OpUpdateOne.String() && payload.Operation != ent.OpCreate.String() {
		return nil
	}

	schemaType := mutationSchemaType(ctx, payload)
	if schemaType == "" {
		return nil
	}

	eventType := workflowEventTypeFromEntOperation(payload.Operation)

	changedFields := lo.Uniq(append([]string(nil), payload.ChangedFields...))
	if eventType != "CREATE" {
		if objectType := enums.ToWorkflowObjectType(schemaType); objectType != nil {
			eligible := workflows.EligibleWorkflowFields(*objectType)
			if len(eligible) > 0 {
				changedFields = lo.Filter(changedFields, func(field string, _ int) bool {
					_, ok := eligible[field]

					return ok
				})
			}
		}
	}

	changedEdges := lo.Uniq(append([]string(nil), payload.ChangedEdges...))
	addedIDs := events.CloneStringSliceMap(payload.AddedIDs)
	removedIDs := events.CloneStringSliceMap(payload.RemovedIDs)
	proposedChanges := events.CloneAnyMap(payload.ProposedChanges)

	if len(changedFields) == 0 && len(changedEdges) == 0 && eventType != "CREATE" {
		return nil
	}

	client := l.client
	if payload.Client != nil {
		client = payload.Client
	}

	// Convert events.MutationPayload to workflows.MutationPayload
	wfPayload := &workflows.MutationPayload{
		Mutation:        payload.Mutation,
		MutationType:    payload.MutationType,
		Operation:       payload.Operation,
		EntityID:        payload.EntityID,
		ChangedFields:   changedFields,
		ClearedFields:   append([]string(nil), payload.ClearedFields...),
		ChangedEdges:    changedEdges,
		AddedIDs:        addedIDs,
		RemovedIDs:      removedIDs,
		ProposedChanges: proposedChanges,
		Client:          payload.Client,
	}
	entityID, ok := workflows.MutationEntityID(ctx, wfPayload)
	if !ok {
		return nil
	}

	allowCtx := workflows.AllowContext(ctx.Context())
	obj, err := loadWorkflowObject(allowCtx, client, schemaType, entityID)
	if err != nil {
		return nil
	}

	definitions, err := l.engine.FindMatchingDefinitions(allowCtx, schemaType, eventType, changedFields, changedEdges, addedIDs, removedIDs, proposedChanges, obj)
	if err != nil || len(definitions) == 0 {
		return nil
	}

	for _, def := range definitions {
		if workflows.DefinitionUsesPreCommitApprovals(def.DefinitionJSON) {
			continue
		}

		_, err := l.engine.TriggerWorkflow(ctx.Context(), def, obj, TriggerInput{
			EventType:       eventType,
			ChangedFields:   changedFields,
			ChangedEdges:    changedEdges,
			AddedIDs:        addedIDs,
			RemovedIDs:      removedIDs,
			ProposedChanges: proposedChanges,
		})
		if err != nil && !errors.Is(err, workflows.ErrWorkflowAlreadyActive) {
			log.Ctx(ctx.Context()).Error().Err(err).Str("definition_id", def.ID).Msg("failed to trigger workflow")
		}
	}

	return nil
}

// HandleWorkflowAssignmentMutation reacts to assignment status changes and emits completion events.
func (l *WorkflowListeners) HandleWorkflowAssignmentMutation(ctx *soiree.EventContext, payload *events.MutationPayload) error {
	if payload == nil {
		return nil
	}

	if payload.Operation != ent.OpUpdate.String() && payload.Operation != ent.OpUpdateOne.String() {
		return nil
	}

	assignmentID := strings.TrimSpace(payload.EntityID)
	if assignmentID == "" {
		if id, ok := ctx.PropertyString("ID"); ok {
			assignmentID = strings.TrimSpace(id)
		}
	}
	if assignmentID == "" {
		return nil
	}

	var (
		newStatus enums.WorkflowAssignmentStatus
		ok        bool
	)

	if mut, typed := payload.Mutation.(*generated.WorkflowAssignmentMutation); typed && mut != nil {
		newStatus, ok = mut.Status()
		if !ok || newStatus == enums.WorkflowAssignmentStatusPending {
			return nil
		}

		oldStatus, err := mut.OldStatus(ctx.Context())
		if err != nil || oldStatus == newStatus {
			return err
		}

		log.Info().Str("assignment_id", assignmentID).Str("old_status", oldStatus.String()).Str("new_status", newStatus.String()).Msg("workflow assignment status changed")
	} else {
		if !lo.Contains(payload.ChangedFields, workflowassignment.FieldStatus) {
			return nil
		}

		rawStatus, found := payload.ProposedChanges[workflowassignment.FieldStatus]
		if !found {
			if status, ok := ctx.PropertyString(workflowassignment.FieldStatus); ok {
				rawStatus = status
				found = true
			}
		}
		if !found {
			return nil
		}

		next := parseWorkflowAssignmentStatus(rawStatus)
		if next == nil || *next == enums.WorkflowAssignmentStatusPending {
			return nil
		}

		newStatus = *next
	}

	return l.engine.CompleteAssignment(ctx.Context(), assignmentID, newStatus, nil, nil)
}

// mutationSchemaType resolves schema type from payload metadata, event properties, or topic fallback.
func mutationSchemaType(ctx *soiree.EventContext, payload *events.MutationPayload) string {
	if schemaType := events.MutationType(payload); schemaType != "" {
		return schemaType
	}

	if schemaType, ok := ctx.PropertyString("mutation_type"); ok {
		if value := strings.TrimSpace(schemaType); value != "" {
			return value
		}
	}

	if ctx != nil {
		if event := ctx.Event(); event != nil {
			if topic := strings.TrimSpace(event.Topic()); topic != "" {
				return topic
			}
		}
	}

	return ""
}

// parseWorkflowAssignmentStatus parses status values from mutation payloads/properties.
func parseWorkflowAssignmentStatus(value any) *enums.WorkflowAssignmentStatus {
	return events.ParseEnumPtr(value, enums.ToWorkflowAssignmentStatus)
}

// loadWorkflowObject loads a workflow object from the generated registry.
func loadWorkflowObject(ctx context.Context, client *generated.Client, schemaType, entityID string) (*workflows.Object, error) {
	entity, err := workflows.LoadWorkflowObject(ctx, client, schemaType, entityID)
	if err != nil {
		return nil, err
	}

	objectType := enums.ToWorkflowObjectType(schemaType)
	if objectType == nil {
		return nil, workflows.ErrUnsupportedObjectType
	}

	return &workflows.Object{
		ID:   entityID,
		Type: *objectType,
		Node: entity,
	}, nil
}

// workflowEventTypeFromEntOperation maps ent mutation ops to workflow event types.
func workflowEventTypeFromEntOperation(operation string) string {
	switch operation {
	case ent.OpUpdate.String(), ent.OpUpdateOne.String():
		return "UPDATE"
	case ent.OpCreate.String():
		return "CREATE"
	case ent.OpDelete.String(), ent.OpDeleteOne.String():
		return "DELETE"
	default:
		return operation
	}
}

// HandleWorkflowTriggered processes a newly triggered workflow instance.
// For approval actions with When conditions, all matching actions are started concurrently.
// This enables workflows where multiple fields change in a single mutation to trigger
// their respective approval requirements simultaneously.
func (l *WorkflowListeners) HandleWorkflowTriggered(ctx *soiree.EventContext, payload soiree.WorkflowTriggeredPayload) (err error) {
	scope := observability.BeginListenerTopic(ctx, l.observer, soiree.WorkflowTriggeredTopic, payload, nil)
	scopeCtx := scope.Context()
	defer scope.End(err, nil)

	instance, orgID, err := l.loadInstanceForScope(scope, payload.InstanceID)
	if err != nil {
		return scope.Fail(err, nil)
	}

	l.recordEvent(scope, instance, enums.WorkflowEventTypeInstanceTriggered, "", payload)

	def := instance.DefinitionSnapshot
	obj := workflowObjectFromPayload(payload.ObjectID, payload.ObjectType)

	if len(def.Actions) == 0 {
		l.emitInstanceCompleted(scope, instance, enums.WorkflowInstanceStateCompleted, obj)
		return nil
	}

	// Find all gated actions with When conditions and start those that match concurrently
	type gatedStart struct {
		action     models.WorkflowAction
		index      int
		obj        *workflows.Object
		actionType enums.WorkflowActionType
	}

	gatedToStart := make([]gatedStart, 0)
	for i, action := range def.Actions {
		actionType := enums.ToWorkflowActionType(action.Type)
		if actionType == nil || !isGatedActionType(*actionType) || action.When == "" {
			continue
		}

		// Evaluate the When condition using existing evaluation logic
		loadedObj, shouldExecute, evalErr := l.evaluateActionWhen(scopeCtx, instance, action, obj, orgID)
		if evalErr != nil || !shouldExecute {
			continue
		}

		gatedToStart = append(gatedToStart, gatedStart{
			action:     action,
			index:      i,
			obj:        loadedObj,
			actionType: *actionType,
		})
	}

	if len(gatedToStart) > 0 {
		keys := lo.Map(gatedToStart, func(item gatedStart, _ int) string {
			return item.action.Key
		})
		keys = workflows.NormalizeStrings(keys)
		if len(keys) > 0 {
			allowCtx := workflows.AllowContext(scopeCtx)
			contextData := instance.Context
			// ParallelApprovalKeys includes review actions as well.
			contextData.ParallelApprovalKeys = keys
			if err := l.client.WorkflowInstance.UpdateOneID(instance.ID).
				SetContext(contextData).
				Exec(allowCtx); err != nil {
				return scope.Fail(err, nil)
			}
		}

		lo.ForEach(gatedToStart, func(start gatedStart, _ int) {
			l.emitActionStarted(scope, instance, start.action.Key, start.index, start.actionType, start.obj)
		})

		return nil
	}

	// If no conditional approvals matched, start the first action normally
	actionType := enums.ToWorkflowActionType(def.Actions[0].Type)
	if actionType != nil {
		l.emitActionStarted(scope, instance, def.Actions[0].Key, 0, *actionType, obj)
	}

	return nil
}

// HandleActionStarted executes a workflow action.
func (l *WorkflowListeners) HandleActionStarted(ctx *soiree.EventContext, payload soiree.WorkflowActionStartedPayload) (err error) {
	scope := observability.BeginListenerTopic(ctx, l.observer, soiree.WorkflowActionStartedTopic, payload, nil)
	scopeCtx := scope.Context()
	defer scope.End(err, nil)

	instance, orgID, err := l.loadInstanceForScope(scope, payload.InstanceID)
	if err != nil {
		return scope.Fail(err, nil)
	}

	def := instance.DefinitionSnapshot
	obj := workflowObjectFromPayload(payload.ObjectID, payload.ObjectType)

	if payload.ActionIndex >= len(def.Actions) {
		scope.RecordError(ErrActionIndexOutOfBounds, nil)
		l.recordActionFailure(scope, instance, actionIndexOutOfBoundsDetails(payload), obj)
		return nil
	}

	action := def.Actions[payload.ActionIndex]
	scope.WithFields(observability.ActionFields(action.Key, nil))

	allowCtx := workflows.AllowContext(scopeCtx)
	if err := l.client.WorkflowInstance.UpdateOneID(instance.ID).
		SetCurrentActionIndex(payload.ActionIndex).
		Exec(allowCtx); err != nil {
		l.recordActionFailure(scope, instance, actionFailureDetails(action.Key, payload.ActionIndex, payload.ActionType, obj, err), obj)
		return scope.Fail(err, observability.Fields{
			"action_index": payload.ActionIndex,
		})
	}

	// Evaluate optional when expression
	var shouldExecute bool
	obj, shouldExecute, err = l.evaluateActionWhen(scopeCtx, instance, action, obj, orgID)
	if err != nil {
		l.recordActionFailure(scope, instance, actionFailureDetails(action.Key, payload.ActionIndex, payload.ActionType, obj, err), obj)
		return scope.Fail(err, nil)
	}

	if !shouldExecute {
		l.skipAction(scope, instance, action, payload, obj)
		return nil
	}

	// Execute action
	execErr := l.engine.ProcessAction(scopeCtx, instance, action)
	if errors.Is(execErr, ErrIntegrationActionQueued) {
		return nil
	}
	if errors.Is(execErr, ErrApprovalNoTargets) || errors.Is(execErr, ErrReviewNoTargets) {
		if err := l.removeParallelApprovalKey(scopeCtx, instance, action.Key); err != nil {
			scope.RecordError(err, nil)
		}
		l.skipAction(scope, instance, action, payload, obj)
		return nil
	}
	if execErr != nil {
		scope.RecordError(execErr, nil)
	}

	actionType := enums.ToWorkflowActionType(action.Type)
	if actionType != nil {
		l.emitActionCompleted(scope, instance, action.Key, payload.ActionIndex, *actionType, obj, execErr, false)
	}

	return nil
}

// HandleActionCompleted determines next steps after action completion.
func (l *WorkflowListeners) HandleActionCompleted(ctx *soiree.EventContext, payload soiree.WorkflowActionCompletedPayload) (err error) {
	scope := observability.BeginListenerTopic(ctx, l.observer, soiree.WorkflowActionCompletedTopic, payload, nil)
	defer scope.End(err, nil)

	instance, orgID, err := l.loadInstanceForScope(scope, payload.InstanceID)
	if err != nil {
		return scope.Fail(err, nil)
	}

	def := instance.DefinitionSnapshot
	obj := workflowObjectFromPayload(payload.ObjectID, payload.ObjectType)

	actionKey := actionKeyForIndex(def.Actions, payload.ActionIndex)
	if actionKey != "" {
		scope.WithFields(observability.ActionFields(actionKey, nil))
	}

	recordable := !isGatedActionType(payload.ActionType) || payload.Skipped || !payload.Success
	if recordable {
		l.recordActionResult(scope, instance, actionKey, payload)
	}

	if !payload.Success {
		scope.RecordError(ErrActionExecutionFailed, nil)
		return l.failInstance(scope, instance, obj, nil, nil)
	}

	if isGatedActionType(payload.ActionType) && payload.Skipped {
		if len(instance.Context.ParallelApprovalKeys) > 0 {
			return nil
		}

		hasRemainingGated := false
		for i := payload.ActionIndex + 1; i < len(def.Actions); i++ {
			actionType := enums.ToWorkflowActionType(def.Actions[i].Type)
			if actionType != nil && isGatedActionType(*actionType) {
				hasRemainingGated = true
				break
			}
		}
		if !hasRemainingGated && instance.WorkflowProposalID != "" {
			if err := l.engine.proposalManager.Apply(scope, instance.WorkflowProposalID, obj); err != nil {
				return l.failInstance(scope, instance, obj, err, observability.Fields{
					workflowinstance.FieldWorkflowProposalID: instance.WorkflowProposalID,
				})
			}
		}
	}

	// If approval action, workflow pauses (listener on assignment completed will resume).
	// Skipped approval actions should advance immediately.
	if isGatedActionType(payload.ActionType) && !payload.Skipped {
		return nil
	}

	nextIndex := payload.ActionIndex + 1
	if err := l.advanceWorkflow(scope, instance, orgID, def, nextIndex, obj); err != nil {
		details := actionCompletedDetailsFromPayload(actionKey, payload)
		details.Success = false
		details.ErrorMessage = err.Error()
		l.recordActionFailure(scope, instance, details, obj)
		return scope.Fail(err, observability.Fields{
			workflowinstance.FieldCurrentActionIndex: nextIndex,
		})
	}

	return nil
}

// HandleAssignmentCompleted handles approval decisions and continues/cancels workflows.
func (l *WorkflowListeners) HandleAssignmentCompleted(ctx *soiree.EventContext, payload soiree.WorkflowAssignmentCompletedPayload) (err error) {
	scope := observability.BeginListenerTopic(ctx, l.observer, soiree.WorkflowAssignmentCompletedTopic, payload, nil)
	scopeCtx := scope.Context()
	defer scope.End(err, nil)

	orgID, err := auth.GetOrganizationIDFromContext(scopeCtx)
	if err != nil {
		return scope.Fail(err, nil)
	}

	assignment, err := l.client.WorkflowAssignment.Query().
		Where(
			workflowassignment.IDEQ(payload.AssignmentID),
			workflowassignment.OwnerIDEQ(orgID),
		).
		Only(scopeCtx)
	if err != nil {
		return scope.Fail(err, nil)
	}

	scope.WithFields(observability.Fields{
		workflowevent.FieldWorkflowInstanceID: assignment.WorkflowInstanceID,
		workflowassignment.FieldStatus:        assignment.Status.String(),
	})

	if isChangeRequestAssignment(assignment) {
		scope.Skip("change_request_assignment", observability.Fields{
			workflowassignment.FieldAssignmentKey: assignment.AssignmentKey,
		})
		return nil
	}

	instance, err := loadWorkflowInstance(scopeCtx, l.client, assignment.WorkflowInstanceID, orgID)
	if err != nil {
		return scope.Fail(err, nil)
	}

	scope.WithFields(observability.Fields{
		workflowinstance.FieldWorkflowDefinitionID: instance.WorkflowDefinitionID,
	})

	def := instance.DefinitionSnapshot

	obj, err := l.loadActionObject(scopeCtx, instance.ID, orgID)
	if err != nil {
		return scope.Fail(err, nil)
	}

	// Re-evaluate NOTIFY actions that may depend on assignment state
	l.reEvaluateNotifyActions(scope, instance, obj)

	// Only resume workflows that are paused waiting for approvals.
	if instance.State != enums.WorkflowInstanceStatePaused {
		scope.Skip("instance_not_paused", observability.Fields{
			workflowinstance.FieldState: instance.State.String(),
		})
		return nil
	}

	// determine which action this assignment corresponds to
	approvalMeta := assignment.ApprovalMetadata
	actionIndex := assignmentActionIndex(def.Actions, assignment.AssignmentKey, approvalMeta.ActionKey)

	if actionIndex == -1 {
		var errFields observability.Fields
		if approvalMeta.ActionKey != "" {
			errFields = observability.ActionFields(approvalMeta.ActionKey, nil)
		}
		scope.RecordError(ErrAssignmentActionNotFound, errFields)
		return nil
	}

	scope.WithFields(observability.ActionFields(def.Actions[actionIndex].Key, nil))

	action := def.Actions[actionIndex]

	allAssignments, err := l.client.WorkflowAssignment.Query().
		Where(
			workflowassignment.WorkflowInstanceIDEQ(instance.ID),
			workflowassignment.OwnerIDEQ(orgID),
		).
		All(scopeCtx)
	if err != nil {
		return scope.Fail(err, nil)
	}

	expectedIndices, useExpected := resolveExpectedActionIndices(def.Actions, instance.Context.ParallelApprovalKeys, actionIndex)
	assignmentsByAction, maxActionIndex := groupAssignmentsByAction(def.Actions, allAssignments, expectedIndices, useExpected, actionIndex)
	assignments := assignmentsByAction[actionIndex]
	if len(assignments) == 0 {
		scope.RecordError(ErrAssignmentActionNotFound, observability.ActionFields(action.Key, nil))
		return nil
	}

	requiredCount := requiredApprovalCount(action, approvalMeta, assignment.Required)

	statusCounts := CountAssignmentStatus(assignments)
	resolution := resolveApproval(requiredCount, statusCounts)
	assignmentIDs := collectAssignmentIDs(assignments)

	targetUserIDs, targetErr := assignmentTargetUserIDs(scopeCtx, l.client, assignmentIDs, orgID)
	if targetErr != nil {
		scope.Warn(targetErr, observability.Fields{
			workflowevent.FieldWorkflowInstanceID: instance.ID,
		})
	}

	if assignment.Status == enums.WorkflowAssignmentStatusChangesRequested {
		if err := l.createChangeRequestAssignment(scope, instance, assignment, action, obj, orgID); err != nil {
			scope.Warn(err, observability.Fields{
				workflowassignment.FieldAssignmentKey: assignment.AssignmentKey,
			})
		}
	}

	if assignment.Status == enums.WorkflowAssignmentStatusChangesRequested {
		if err := l.closePendingApprovalsForChangeRequest(scope, assignmentsByAction, expectedIndices, actionIndex, useExpected, assignment, approvalMeta.ActionKey); err != nil {
			scope.Warn(err, observability.Fields{
				workflowassignment.FieldAssignmentKey: assignment.AssignmentKey,
			})
		}

		label := approvalMeta.Label
		l.recordApprovalEvent(scope, instance, action, actionIndex, obj, statusCounts, requiredCount, assignmentIDs, targetUserIDs, false, label, assignment.Required)

		return nil
	}

	if resolution != approvalPending {
		label := approvalMeta.Label
		l.recordApprovalEvent(scope, instance, action, actionIndex, obj, statusCounts, requiredCount, assignmentIDs, targetUserIDs, resolution == approvalSatisfied, label, assignment.Required)
	}

	switch resolution {
	case approvalFailed:
		if statusCounts.RejectedRequired || statusCounts.ChangesRequestedRequired {
			scope.Info(observability.ActionFields(def.Actions[actionIndex].Key, nil))
		}
		return l.failInstance(scope, instance, obj, nil, nil)
	case approvalPending:
		return nil
	}

	allResolved, failed := evaluateApprovalGroups(def.Actions, assignmentsByAction, expectedIndices, useExpected)
	if failed {
		return l.failInstance(scope, instance, obj, nil, nil)
	}

	if !allResolved {
		return nil
	}

	// Apply proposed changes once approvals satisfy quorum/all required targets.
	if instance.WorkflowProposalID != "" {
		if err := l.engine.proposalManager.Apply(scope, instance.WorkflowProposalID, obj); err != nil {
			return l.failInstance(scope, instance, obj, err, observability.Fields{
				workflowinstance.FieldWorkflowProposalID: instance.WorkflowProposalID,
			})
		}
	}

	// Resume workflow by re-entering the normal action pipeline using compare-and-swap to prevent races.
	clearParallel := len(expectedIndices) > 0
	if err := l.resumeWorkflowAfterApproval(scope, instance, orgID, maxActionIndex, obj, def, clearParallel); err != nil {
		return l.failInstance(scope, instance, obj, err, nil)
	}

	return nil
}

// isChangeRequestAssignment reports whether an assignment is the change-request leg.
func isChangeRequestAssignment(assignment *generated.WorkflowAssignment) bool {
	return strings.HasPrefix(assignment.AssignmentKey, "change_request_") || assignment.Role == "REQUESTER"
}

// resolveExpectedActionIndices resolves action indexes from configured parallel action keys.
func resolveExpectedActionIndices(actions []models.WorkflowAction, parallelKeys []string, actionIndex int) (map[int]struct{}, bool) {
	expectedKeys := workflows.NormalizeStrings(parallelKeys)
	indexes := lo.FilterMap(expectedKeys, func(key string, _ int) (int, bool) {
		idx := actionIndexForKey(actions, key)
		return idx, idx >= 0
	})
	expectedIndices := lo.SliceToMap(indexes, func(idx int) (int, struct{}) {
		return idx, struct{}{}
	})

	useExpected := len(expectedIndices) > 0
	if useExpected {
		if _, ok := expectedIndices[actionIndex]; !ok {
			useExpected = false
		}
	}

	return expectedIndices, useExpected
}

// groupAssignmentsByAction buckets assignments by action index for approval evaluation.
func groupAssignmentsByAction(actions []models.WorkflowAction, allAssignments []*generated.WorkflowAssignment, expectedIndices map[int]struct{}, useExpected bool, actionIndex int) (map[int][]*generated.WorkflowAssignment, int) {
	assignmentsByAction := make(map[int][]*generated.WorkflowAssignment)
	maxActionIndex := -1

	for _, a := range allAssignments {
		idx := assignmentActionIndex(actions, a.AssignmentKey, a.ApprovalMetadata.ActionKey)
		if idx == -1 {
			continue
		}
		if useExpected {
			if _, ok := expectedIndices[idx]; !ok {
				continue
			}
		}
		assignmentsByAction[idx] = append(assignmentsByAction[idx], a)
		if idx > maxActionIndex {
			maxActionIndex = idx
		}
	}

	if useExpected {
		for idx := range expectedIndices {
			if idx > maxActionIndex {
				maxActionIndex = idx
			}
		}
	} else if maxActionIndex == -1 {
		maxActionIndex = actionIndex
	}

	return assignmentsByAction, maxActionIndex
}

// collectAssignmentIDs returns assignment IDs for event metadata payloads.
func collectAssignmentIDs(assignments []*generated.WorkflowAssignment) []string {
	return lo.Map(assignments, func(a *generated.WorkflowAssignment, _ int) string { return a.ID })
}

// recordApprovalEvent emits normalized approval/review completion event details.
func (l *WorkflowListeners) recordApprovalEvent(
	scope *observability.Scope,
	instance *generated.WorkflowInstance,
	action models.WorkflowAction,
	actionIndex int,
	obj *workflows.Object,
	statusCounts AssignmentStatusCounts,
	requiredCount int,
	assignmentIDs []string,
	targetUserIDs []string,
	satisfied bool,
	label string,
	required bool,
) {
	actionType := enums.ToWorkflowActionType(action.Type)
	if actionType != nil && *actionType == enums.WorkflowActionTypeReview {
		details := reviewCompletedDetails(action, actionIndex, obj, statusCounts, requiredCount, assignmentIDs, targetUserIDs, satisfied, label, required)
		l.recordEvent(scope, instance, enums.WorkflowEventTypeActionCompleted, action.Key, details)
		return
	}

	details := approvalCompletedDetails(action, actionIndex, obj, statusCounts, requiredCount, assignmentIDs, targetUserIDs, satisfied, label, required)
	l.recordEvent(scope, instance, enums.WorkflowEventTypeActionCompleted, action.Key, details)
}

// evaluateApprovalGroups evaluates grouped assignments and reports resolution/failure status.
func evaluateApprovalGroups(actions []models.WorkflowAction, assignmentsByAction map[int][]*generated.WorkflowAssignment, expectedIndices map[int]struct{}, useExpected bool) (bool, bool) {
	allResolved := true

	checkGroup := func(idx int, grouped []*generated.WorkflowAssignment) (bool, bool) {
		if len(grouped) == 0 {
			return false, false
		}
		groupMeta := grouped[0].ApprovalMetadata
		groupRequired := requiredApprovalCount(actions[idx], groupMeta, grouped[0].Required)
		groupCounts := CountAssignmentStatus(grouped)
		groupResolution := resolveApproval(groupRequired, groupCounts)

		if groupResolution == approvalFailed {
			return false, true
		}

		if groupResolution == approvalPending {
			return false, false
		}

		return true, false
	}

	if useExpected {
		for idx := range expectedIndices {
			resolved, failed := checkGroup(idx, assignmentsByAction[idx])
			if failed {
				return false, true
			}
			if !resolved {
				allResolved = false
			}
		}
		return allResolved, false
	}

	for idx, grouped := range assignmentsByAction {
		resolved, failed := checkGroup(idx, grouped)
		if failed {
			return false, true
		}
		if !resolved {
			allResolved = false
		}
	}

	return allResolved, false
}

// closePendingApprovalsForChangeRequest closes any outstanding approvals when theres a change request submitted
func (l *WorkflowListeners) closePendingApprovalsForChangeRequest(scope *observability.Scope, assignmentsByAction map[int][]*generated.WorkflowAssignment, expectedIndices map[int]struct{}, actionIndex int, useExpected bool, requesterAssignment *generated.WorkflowAssignment, actionKey string) error {
	if requesterAssignment == nil {
		return nil
	}

	allowCtx := workflows.AllowContext(scope.Context())
	requesterID := requesterAssignment.ActorUserID
	decidedAt := time.Now().UTC()

	indices := make([]int, 0, len(expectedIndices))
	if useExpected {
		indices = lo.Keys(expectedIndices)
	} else {
		indices = append(indices, actionIndex)
	}

	for _, idx := range indices {
		assignments := assignmentsByAction[idx]
		for _, a := range assignments {
			if a == nil || a.ID == requesterAssignment.ID {
				continue
			}
			if a.Status != enums.WorkflowAssignmentStatusPending {
				continue
			}

			rejection := models.WorkflowAssignmentRejection{
				ActionKey:        actionKey,
				RejectionReason:  "changes requested",
				RejectedAt:       decidedAt.Format(time.RFC3339),
				RejectedByUserID: requesterID,
			}
			if rejection.ActionKey == "" {
				rejection.ActionKey = a.ApprovalMetadata.ActionKey
			}

			update := l.client.WorkflowAssignment.UpdateOneID(a.ID).
				SetStatus(enums.WorkflowAssignmentStatusRejected).
				SetRejectionMetadata(rejection).
				SetDecidedAt(decidedAt)

			if requesterID != "" {
				update.SetActorUserID(requesterID)
			}

			if err := update.Exec(allowCtx); err != nil {
				return err
			}

			a.Status = enums.WorkflowAssignmentStatusRejected
		}
	}

	return nil
}

// HandleInstanceCompleted marks a workflow instance as completed or failed.
func (l *WorkflowListeners) HandleInstanceCompleted(ctx *soiree.EventContext, payload soiree.WorkflowInstanceCompletedPayload) (err error) {
	scope := observability.BeginListenerTopic(ctx, l.observer, soiree.WorkflowInstanceCompletedTopic, payload, nil)
	scopeCtx := scope.Context()
	defer scope.End(err, nil)

	instance, orgID, err := l.loadInstanceForScope(scope, payload.InstanceID)
	if err != nil {
		return scope.Fail(err, nil)
	}

	l.recordEvent(scope, instance, enums.WorkflowEventTypeInstanceCompleted, "", payload)

	// Use allow context to avoid GraphQL/FGA edit checks blocking system completion
	// while preserving the original context (organization, client, etc.)
	allowCtx := workflows.AllowContext(scopeCtx)

	// Use compare-and-swap to prevent double-completion races
	updated, err := l.client.WorkflowInstance.Update().
		Where(
			workflowinstance.IDEQ(payload.InstanceID),
			workflowinstance.Not(workflowinstance.StateIn(enums.WorkflowInstanceStateCompleted, enums.WorkflowInstanceStateFailed)),
			workflowinstance.OwnerIDEQ(orgID),
		).
		SetState(payload.State).
		Save(allowCtx)
	if err != nil {
		return scope.Fail(err, nil)
	}

	if updated == 0 {
		scope.Skip("instance_already_terminal", observability.Fields{
			workflowinstance.FieldState: instance.State.String(),
		})

		return nil
	}

	return nil
}

// HandleAssignmentCreated records audit events for assignment creation.
func (l *WorkflowListeners) HandleAssignmentCreated(ctx *soiree.EventContext, payload soiree.WorkflowAssignmentCreatedPayload) (err error) {
	scope := observability.BeginListenerTopic(ctx, l.observer, soiree.WorkflowAssignmentCreatedTopic, payload, nil)
	defer scope.End(err, nil)

	return nil
}

// resumeWorkflowAfterApproval advances a paused workflow after approvals complete
func (l *WorkflowListeners) resumeWorkflowAfterApproval(scope *observability.Scope, instance *generated.WorkflowInstance, orgID string, actionIndex int, obj *workflows.Object, def models.WorkflowDefinitionDocument, clearParallel bool) error {
	allowCtx := workflows.AllowContext(scope.Context())
	update := l.client.WorkflowInstance.Update().
		Where(
			workflowinstance.IDEQ(instance.ID),
			workflowinstance.StateEQ(enums.WorkflowInstanceStatePaused),
			workflowinstance.OwnerIDEQ(orgID),
		).
		SetState(enums.WorkflowInstanceStateRunning).
		SetCurrentActionIndex(actionIndex + 1)
	if clearParallel {
		contextData := instance.Context
		contextData.ParallelApprovalKeys = nil
		update = update.SetContext(contextData)
	}
	updated, err := update.Save(allowCtx)
	if err != nil {
		return err
	}
	if updated == 0 {
		scope.Skip("instance_not_paused", observability.Fields{
			workflowinstance.FieldState: instance.State.String(),
		})
		return nil
	}

	nextIndex := actionIndex + 1
	if nextIndex < len(def.Actions) {
		actionType := enums.ToWorkflowActionType(def.Actions[nextIndex].Type)
		if actionType != nil {
			l.emitActionStarted(scope, instance, def.Actions[nextIndex].Key, nextIndex, *actionType, obj)
		}
		return nil
	}

	l.emitInstanceCompleted(scope, instance, enums.WorkflowInstanceStateCompleted, obj)
	return nil
}

// actionKeyForIndex resolves an action key safely by index
func actionKeyForIndex(actions []models.WorkflowAction, index int) string {
	if index < 0 || index >= len(actions) {
		return ""
	}
	return actions[index].Key
}

// advanceWorkflow updates the current index and emits the next action or completion event
func (l *WorkflowListeners) advanceWorkflow(scope *observability.Scope, instance *generated.WorkflowInstance, orgID string, def models.WorkflowDefinitionDocument, nextIndex int, obj *workflows.Object) error {
	allowCtx := workflows.AllowContext(scope.Context())
	if err := l.client.WorkflowInstance.Update().
		Where(
			workflowinstance.IDEQ(instance.ID),
			workflowinstance.OwnerIDEQ(orgID),
		).
		SetCurrentActionIndex(nextIndex).
		Exec(allowCtx); err != nil {
		return err
	}

	if nextIndex < len(def.Actions) {
		actionType := enums.ToWorkflowActionType(def.Actions[nextIndex].Type)
		if actionType != nil {
			l.emitActionStarted(scope, instance, def.Actions[nextIndex].Key, nextIndex, *actionType, obj)
		}
		return nil
	}

	l.emitInstanceCompleted(scope, instance, enums.WorkflowInstanceStateCompleted, obj)
	return nil
}

// assignmentActionIndex returns the action index associated with a gated assignment
func assignmentActionIndex(actions []models.WorkflowAction, assignmentKey, actionKey string) int {
	_, index, ok := lo.FindIndexOf(actions, func(action models.WorkflowAction) bool {
		if strings.HasPrefix(assignmentKey, fmt.Sprintf("approval_%s_", action.Key)) {
			return true
		}
		if strings.HasPrefix(assignmentKey, fmt.Sprintf("review_%s_", action.Key)) {
			return true
		}
		return actionKey != "" && actionKey == action.Key
	})
	if !ok {
		return -1
	}
	return index
}

// createChangeRequestAssignment creates a new assignment for the requester to address change requests
func (l *WorkflowListeners) createChangeRequestAssignment(scope *observability.Scope, instance *generated.WorkflowInstance, assignment *generated.WorkflowAssignment, action models.WorkflowAction, obj *workflows.Object, orgID string) error {
	if instance == nil || assignment == nil {
		return nil
	}

	requesterID := instance.Context.TriggerUserID
	if requesterID == "" {
		return nil
	}

	actionKey := action.Key
	if actionKey == "" {
		actionKey = assignment.ApprovalMetadata.ActionKey
	}
	if actionKey == "" {
		actionKey = "unknown"
	}

	assignmentKey := fmt.Sprintf("change_request_%s_%s", actionKey, requesterID)
	label := "Changes requested"

	metadata := map[string]any{
		"change_request_for_assignment_id": assignment.ID,
		"change_request_action_key":        actionKey,
		"change_request_action_type":       action.Type,
	}
	if assignment.ActorUserID != "" {
		metadata["change_requested_by_user_id"] = assignment.ActorUserID
	}
	if assignment.RejectionMetadata.RejectionReason != "" {
		metadata["change_request_reason"] = assignment.RejectionMetadata.RejectionReason
	}
	if len(assignment.RejectionMetadata.ChangeRequestInputs) > 0 {
		metadata["change_request_inputs"] = assignment.RejectionMetadata.ChangeRequestInputs
	}
	if assignment.RejectionMetadata.RejectedAt != "" {
		metadata["change_requested_at"] = assignment.RejectionMetadata.RejectedAt
	}

	allowCtx := workflows.AllowContext(scope.Context())

	create := l.client.WorkflowAssignment.Create().
		SetWorkflowInstanceID(instance.ID).
		SetAssignmentKey(assignmentKey).
		SetStatus(enums.WorkflowAssignmentStatusPending).
		SetRequired(false).
		SetRole("REQUESTER").
		SetLabel(label).
		SetMetadata(metadata)
	create.SetOwnerID(orgID)
	if assignment.RejectionMetadata.RejectionReason != "" {
		create.SetNotes(assignment.RejectionMetadata.RejectionReason)
	}

	assignmentCreated := true
	requesterAssignment, err := create.Save(allowCtx)
	if err != nil && generated.IsConstraintError(err) {
		assignmentCreated = false
		requesterAssignment, err = l.client.WorkflowAssignment.Query().
			Where(
				workflowassignment.WorkflowInstanceIDEQ(instance.ID),
				workflowassignment.AssignmentKeyEQ(assignmentKey),
				workflowassignment.OwnerIDEQ(orgID),
			).
			Only(allowCtx)
	}
	if err != nil {
		return err
	}

	if requesterAssignment != nil && !assignmentCreated {
		update := l.client.WorkflowAssignment.UpdateOneID(requesterAssignment.ID).
			SetStatus(enums.WorkflowAssignmentStatusPending).
			SetRole("REQUESTER").
			SetRequired(false).
			SetLabel(label).
			SetMetadata(metadata).
			ClearDecidedAt().
			ClearActorUserID().
			ClearActorGroupID()
		if assignment.RejectionMetadata.RejectionReason != "" {
			update.SetNotes(assignment.RejectionMetadata.RejectionReason)
		}
		if _, updateErr := update.Save(allowCtx); updateErr != nil {
			scope.Warn(updateErr, observability.Fields{
				workflowassignment.FieldAssignmentKey: assignmentKey,
			})
		}
	}

	if requesterAssignment != nil {
		targetCreate := l.client.WorkflowAssignmentTarget.
			Create().
			SetWorkflowAssignmentID(requesterAssignment.ID).
			SetTargetType(enums.WorkflowTargetTypeUser).
			SetTargetUserID(requesterID).
			SetOwnerID(orgID)
		if err := targetCreate.Exec(allowCtx); err != nil && !generated.IsConstraintError(err) {
			return err
		}
	}

	if assignmentCreated && requesterAssignment != nil && obj != nil {
		actionType := enums.ToWorkflowActionType(action.Type)
		if actionType != nil {
			l.engine.emitAssignmentCreated(allowCtx, instance, obj, requesterAssignment.ID, requesterID, *actionType)
		}
	}

	return nil
}

// requiredApprovalCount resolves the approval quorum requirement for an action
func requiredApprovalCount(action models.WorkflowAction, meta models.WorkflowAssignmentApproval, required bool) int {
	requiredCount := meta.RequiredCount
	if requiredCount == 0 {
		var params struct {
			RequiredCount int `json:"required_count"`
		}
		if err := json.Unmarshal(action.Params, &params); err == nil && params.RequiredCount > 0 {
			requiredCount = params.RequiredCount
		}
	}

	// If approvals are optional (required=false) and no quorum was configured, default to one approval
	if requiredCount == 0 && !required {
		requiredCount = 1
	}

	return requiredCount
}

// loadInstanceForScope loads a workflow instance and annotates scope fields
// Uses AllowContext since all callers are internal workflow operations
func (l *WorkflowListeners) loadInstanceForScope(scope *observability.Scope, instanceID string) (*generated.WorkflowInstance, string, error) {
	ctx := scope.Context()

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return nil, "", err
	}

	allowCtx := workflows.AllowContext(ctx)

	instance, err := loadWorkflowInstance(allowCtx, l.client, instanceID, orgID)
	if err != nil {
		return nil, "", err
	}

	scope.WithFields(observability.Fields{
		workflowinstance.FieldWorkflowDefinitionID: instance.WorkflowDefinitionID,
	})

	return instance, orgID, nil
}

// failInstance records a failed completion and optionally fails the scope
func (l *WorkflowListeners) failInstance(scope *observability.Scope, instance *generated.WorkflowInstance, obj *workflows.Object, err error, fields observability.Fields) error {
	l.emitInstanceCompleted(scope, instance, enums.WorkflowInstanceStateFailed, obj)
	if err == nil {
		return nil
	}
	return scope.Fail(err, fields)
}

// workflowObjectFromPayload builds a workflow object from payload identifiers
func workflowObjectFromPayload(objectID string, objectType enums.WorkflowObjectType) *workflows.Object {
	return &workflows.Object{
		ID:   objectID,
		Type: objectType,
	}
}

// skipAction records a skipped action and emits completion when possible
func (l *WorkflowListeners) skipAction(scope *observability.Scope, instance *generated.WorkflowInstance, action models.WorkflowAction, payload soiree.WorkflowActionStartedPayload, obj *workflows.Object) {
	actionType := enums.ToWorkflowActionType(action.Type)
	if actionType != nil {
		l.emitActionCompleted(scope, instance, action.Key, payload.ActionIndex, *actionType, obj, nil, true)
	}
}

// recordActionResult records the completed or failed action event
func (l *WorkflowListeners) recordActionResult(scope *observability.Scope, instance *generated.WorkflowInstance, actionKey string, payload soiree.WorkflowActionCompletedPayload) {
	details := actionCompletedDetailsFromPayload(actionKey, payload)
	l.recordEvent(scope, instance, enums.WorkflowEventTypeActionCompleted, actionKey, details)
}

// removeParallelApprovalKey removes an action key from the parallel approval context list.
func (l *WorkflowListeners) removeParallelApprovalKey(ctx context.Context, instance *generated.WorkflowInstance, actionKey string) error {
	if instance == nil || actionKey == "" {
		return nil
	}

	keys := workflows.NormalizeStrings(instance.Context.ParallelApprovalKeys)
	if len(keys) == 0 {
		return nil
	}

	updatedKeys := lo.Filter(keys, func(key string, _ int) bool {
		return key != actionKey
	})
	if len(updatedKeys) == len(keys) {
		return nil
	}

	allowCtx := workflows.AllowContext(ctx)
	contextData := instance.Context
	contextData.ParallelApprovalKeys = updatedKeys
	if err := l.client.WorkflowInstance.UpdateOneID(instance.ID).
		SetContext(contextData).
		Exec(allowCtx); err != nil {
		return err
	}

	instance.Context = contextData
	return nil
}
