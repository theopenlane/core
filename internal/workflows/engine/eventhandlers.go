package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

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
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/internal/ent/workflowgenerated"
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
	if workflows.IsWorkflowBypass(ctx.Context()) {
		return nil
	}

	if payload.Operation != ent.OpUpdate.String() && payload.Operation != ent.OpUpdateOne.String() {
		return nil
	}

	mut, ok := payload.Mutation.(utils.GenericMutation)
	if !ok {
		return nil
	}

	changedFields := workflows.CollectChangedFields(mut)
	changedEdges, addedIDs, removedIDs := workflowgenerated.ExtractChangedEdges(payload.Mutation)

	if len(changedFields) == 0 && len(changedEdges) == 0 {
		return nil
	}

	client := l.client
	if payload.Client != nil {
		client = payload.Client
	}

	// Convert events.MutationPayload to workflows.MutationPayload
	wfPayload := &workflows.MutationPayload{
		Mutation:  payload.Mutation,
		Operation: payload.Operation,
		Client:    payload.Client,
	}
	entityID, ok := workflows.MutationEntityID(ctx, wfPayload)
	if !ok {
		return nil
	}

	obj, err := loadWorkflowObject(ctx.Context(), client, payload.Mutation.Type(), entityID)
	if err != nil {
		return nil
	}

	eventType := workflowEventTypeFromEntOperation(payload.Operation)

	allowCtx := workflows.AllowContext(ctx.Context())
	definitions, err := l.engine.FindMatchingDefinitions(allowCtx, payload.Mutation.Type(), eventType, changedFields, changedEdges, addedIDs, removedIDs, obj)
	if err != nil || len(definitions) == 0 {
		return nil
	}

	proposedChanges := workflows.BuildProposedChanges(mut, changedFields)

	for _, def := range definitions {
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
	mut, ok := payload.Mutation.(*generated.WorkflowAssignmentMutation)
	if !ok {
		return nil
	}

	if payload.Operation != ent.OpUpdate.String() && payload.Operation != ent.OpUpdateOne.String() {
		return nil
	}

	newStatus, ok := mut.Status()
	if !ok || newStatus == enums.WorkflowAssignmentStatusPending {
		return nil
	}

	oldStatus, err := mut.OldStatus(ctx.Context())
	if err != nil || oldStatus == newStatus {
		return err
	}

	assignmentID, _ := mut.ID()
	if assignmentID == "" {
		return nil
	}

	log.Info().Str("assignment_id", assignmentID).Str("old_status", oldStatus.String()).Str("new_status", newStatus.String()).Msg("workflow assignment status changed")

	return l.engine.CompleteAssignment(ctx.Context(), assignmentID, newStatus, nil, nil)
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
func (l *WorkflowListeners) HandleWorkflowTriggered(ctx *soiree.EventContext, payload soiree.WorkflowTriggeredPayload) (err error) {
	scope := observability.BeginListenerTopic(ctx, l.observer, soiree.WorkflowTriggeredTopic, payload, nil)
	defer scope.End(err, nil)

	instance, _, err := l.loadInstanceForScope(scope, payload.InstanceID)
	if err != nil {
		return scope.Fail(err, nil)
	}

	l.recordEvent(scope, instance, enums.WorkflowEventTypeInstanceTriggered, "", payload)

	def := instance.DefinitionSnapshot
	obj := workflowObjectFromPayload(payload.ObjectID, payload.ObjectType)

	// Process first action
	if len(def.Actions) > 0 {
		actionType := enums.ToWorkflowActionType(def.Actions[0].Type)
		if actionType != nil {
			l.emitActionStarted(scope, instance, def.Actions[0].Key, 0, *actionType, obj)
		}
	} else {
		// No actions, mark as completed
		l.emitInstanceCompleted(scope, instance, enums.WorkflowInstanceStateCompleted, obj)
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
	shouldExecute := true
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

	recordable := payload.ActionType != enums.WorkflowActionTypeApproval || payload.Skipped || !payload.Success
	if recordable {
		l.recordActionResult(scope, instance, actionKey, payload)
	}

	if !payload.Success {
		scope.RecordError(ErrActionExecutionFailed, nil)
		return l.failInstance(scope, instance, obj, nil, nil)
	}

	// If approval action, workflow pauses (listener on assignment completed will resume).
	// Skipped approval actions should advance immediately.
	if payload.ActionType == enums.WorkflowActionTypeApproval && !payload.Skipped {
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
	actionIndex := approvalActionIndex(def.Actions, assignment.AssignmentKey, approvalMeta.ActionKey)

	if actionIndex == -1 {
		var errFields observability.Fields
		if approvalMeta.ActionKey != "" {
			errFields = observability.ActionFields(approvalMeta.ActionKey, nil)
		}
		scope.RecordError(ErrAssignmentActionNotFound, errFields)
		return nil
	}

	scope.WithFields(observability.ActionFields(def.Actions[actionIndex].Key, nil))

	prefix := fmt.Sprintf("approval_%s_", def.Actions[actionIndex].Key)
	assignments, err := l.client.WorkflowAssignment.Query().
		Where(
			workflowassignment.WorkflowInstanceIDEQ(instance.ID),
			workflowassignment.AssignmentKeyHasPrefix(prefix),
			workflowassignment.OwnerIDEQ(orgID),
		).All(scopeCtx)
	if err != nil {
		return scope.Fail(err, nil)
	}

	requiredCount := requiredApprovalCount(def.Actions[actionIndex], approvalMeta, assignment.Required)

	statusCounts := CountAssignmentStatus(assignments)
	resolution := resolveApproval(requiredCount, statusCounts)
	assignmentIDs := make([]string, 0, len(assignments))
	for _, a := range assignments {
		assignmentIDs = append(assignmentIDs, a.ID)
	}
	targetUserIDs, targetErr := assignmentTargetUserIDs(scopeCtx, l.client, assignmentIDs, orgID)
	if targetErr != nil {
		scope.Warn(targetErr, observability.Fields{
			workflowevent.FieldWorkflowInstanceID: instance.ID,
		})
	}
	if resolution != approvalPending {
		label := approvalMeta.Label
		details := approvalCompletedDetails(def.Actions[actionIndex], actionIndex, obj, statusCounts, requiredCount, assignmentIDs, targetUserIDs, resolution == approvalSatisfied, label, assignment.Required)
		l.recordEvent(scope, instance, enums.WorkflowEventTypeActionCompleted, def.Actions[actionIndex].Key, details)
	}
	switch resolution {
	case approvalFailed:
		if statusCounts.RejectedRequired {
			scope.Info(observability.ActionFields(def.Actions[actionIndex].Key, nil))
		}
		return l.failInstance(scope, instance, obj, nil, nil)
	case approvalPending:
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
	if err := l.resumeWorkflowAfterApproval(scope, instance, orgID, actionIndex, obj, def); err != nil {
		return l.failInstance(scope, instance, obj, err, nil)
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
func (l *WorkflowListeners) resumeWorkflowAfterApproval(scope *observability.Scope, instance *generated.WorkflowInstance, orgID string, actionIndex int, obj *workflows.Object, def models.WorkflowDefinitionDocument) error {
	allowCtx := workflows.AllowContext(scope.Context())
	updated, err := l.client.WorkflowInstance.Update().
		Where(
			workflowinstance.IDEQ(instance.ID),
			workflowinstance.StateEQ(enums.WorkflowInstanceStatePaused),
			workflowinstance.OwnerIDEQ(orgID),
		).
		SetState(enums.WorkflowInstanceStateRunning).
		SetCurrentActionIndex(actionIndex + 1).
		Save(allowCtx)
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

// approvalActionIndex returns the action index associated with an approval assignment
func approvalActionIndex(actions []models.WorkflowAction, assignmentKey, actionKey string) int {
	_, index, ok := lo.FindIndexOf(actions, func(action models.WorkflowAction) bool {
		if strings.HasPrefix(assignmentKey, fmt.Sprintf("approval_%s_", action.Key)) {
			return true
		}
		return actionKey != "" && actionKey == action.Key
	})
	if !ok {
		return -1
	}
	return index
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
func (l *WorkflowListeners) loadInstanceForScope(scope *observability.Scope, instanceID string) (*generated.WorkflowInstance, string, error) {
	orgID, err := auth.GetOrganizationIDFromContext(scope.Context())
	if err != nil {
		return nil, "", err
	}

	instance, err := loadWorkflowInstance(scope.Context(), l.client, instanceID, orgID)
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
