package engine

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/workflowassignmenttarget"
	"github.com/theopenlane/core/internal/ent/generated/workflowevent"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/internal/workflows/observability"
	"github.com/theopenlane/core/pkg/events/soiree"
)

// emitAssignmentCreated emits assignment created events and records enqueue failures
func (e *WorkflowEngine) emitAssignmentCreated(ctx context.Context, instance *generated.WorkflowInstance, obj *workflows.Object, assignmentID string, userID string) {
	payload := soiree.WorkflowAssignmentCreatedPayload{
		AssignmentID: assignmentID,
		InstanceID:   instance.ID,
		TargetType:   enums.WorkflowTargetTypeUser,
		TargetIDs:    []string{userID},
		ObjectID:     obj.ID,
		ObjectType:   obj.Type,
	}
	meta := workflows.EmitFailureMeta{
		EventType:   enums.WorkflowEventTypeAssignmentCreated,
		ActionKey:   "",
		ActionIndex: -1,
		ObjectID:    obj.ID,
		ObjectType:  obj.Type,
	}
	emitEngineEvent(ctx, e, observability.OpExecuteAction, enums.WorkflowActionTypeApproval.String(), instance, meta, soiree.WorkflowAssignmentCreatedTopic, payload, observability.Fields{
		workflowassignmenttarget.FieldTargetUserID: userID,
	})
}

// emitActionStarted emits action started event
func (l *WorkflowListeners) emitActionStarted(scope *observability.Scope, instance *generated.WorkflowInstance, actionKey string, actionIndex int, actionType enums.WorkflowActionType, obj *workflows.Object) {
	payload := soiree.WorkflowActionStartedPayload{
		InstanceID:  instance.ID,
		ActionIndex: actionIndex,
		ActionType:  actionType,
		ObjectID:    obj.ID,
		ObjectType:  obj.Type,
	}
	meta := workflows.EmitFailureMeta{
		EventType:   enums.WorkflowEventTypeActionStarted,
		ActionKey:   actionKey,
		ActionIndex: actionIndex,
		ObjectID:    obj.ID,
		ObjectType:  obj.Type,
	}
	emitListenerEvent(scope, l, instance, meta, soiree.WorkflowActionStartedTopic, payload)
}

// emitActionCompleted emits action completed event
func (l *WorkflowListeners) emitActionCompleted(scope *observability.Scope, instance *generated.WorkflowInstance, actionKey string, actionIndex int, actionType enums.WorkflowActionType, obj *workflows.Object, execErr error, skipped bool) {
	payload := soiree.WorkflowActionCompletedPayload{
		InstanceID:  instance.ID,
		ActionIndex: actionIndex,
		ActionType:  actionType,
		ObjectID:    obj.ID,
		ObjectType:  obj.Type,
		Success:     execErr == nil,
		Skipped:     skipped,
	}

	if execErr != nil {
		payload.ErrorMessage = execErr.Error()
	}
	meta := workflows.EmitFailureMeta{
		EventType:   enums.WorkflowEventTypeActionCompleted,
		ActionKey:   actionKey,
		ActionIndex: actionIndex,
		ObjectID:    obj.ID,
		ObjectType:  obj.Type,
	}
	emitListenerEvent(scope, l, instance, meta, soiree.WorkflowActionCompletedTopic, payload)
}

// emitInstanceCompleted emits instance completed event
func (l *WorkflowListeners) emitInstanceCompleted(scope *observability.Scope, instance *generated.WorkflowInstance, state enums.WorkflowInstanceState, obj *workflows.Object) {
	payload := soiree.WorkflowInstanceCompletedPayload{
		InstanceID: instance.ID,
		State:      state,
		ObjectID:   obj.ID,
		ObjectType: obj.Type,
	}
	meta := workflows.EmitFailureMeta{
		EventType:   enums.WorkflowEventTypeInstanceCompleted,
		ActionKey:   "",
		ActionIndex: -1,
		ObjectID:    obj.ID,
		ObjectType:  obj.Type,
	}
	emitListenerEvent(scope, l, instance, meta, soiree.WorkflowInstanceCompletedTopic, payload)
}

// emitWorkflowTriggered emits workflow.triggered for a workflow instance
func (e *WorkflowEngine) emitWorkflowTriggered(ctx context.Context, op observability.OperationName, trigger string, instance *generated.WorkflowInstance, defID string, obj *workflows.Object, changedFields []string) {
	payload := soiree.WorkflowTriggeredPayload{
		InstanceID:           instance.ID,
		DefinitionID:         defID,
		ObjectID:             obj.ID,
		ObjectType:           obj.Type,
		TriggerEventType:     trigger,
		TriggerChangedFields: changedFields,
	}
	meta := workflows.EmitFailureMeta{
		EventType:   enums.WorkflowEventTypeInstanceTriggered,
		ActionKey:   "",
		ActionIndex: -1,
		ObjectID:    obj.ID,
		ObjectType:  obj.Type,
	}
	emitEngineEvent(ctx, e, op, trigger, instance, meta, soiree.WorkflowTriggeredTopic, payload, observability.Fields{
		workflowevent.FieldWorkflowInstanceID: instance.ID,
	})
}

// emitEngineEvent emits a typed event and records enqueue failures for engine operations
func emitEngineEvent[T any](ctx context.Context, engine *WorkflowEngine, op observability.OperationName, trigger string, instance *generated.WorkflowInstance, meta workflows.EmitFailureMeta, topic soiree.TypedTopic[T], payload T, warnFields observability.Fields) {
	receipt := workflows.EmitWorkflowEvent(ctx, engine.emitter, topic, payload, engine.client)
	if receipt.Err == nil {
		return
	}

	recordEmitFailure(ctx, engine.client, instance, meta, topic.Name(), payload, receipt)
	observability.WarnEngine(ctx, op, trigger, warnFields, receipt.Err)
}

// emitListenerEvent emits a typed event and records enqueue failures for listener operations
func emitListenerEvent[T any](scope *observability.Scope, listeners *WorkflowListeners, instance *generated.WorkflowInstance, meta workflows.EmitFailureMeta, topic soiree.TypedTopic[T], payload T) {
	receipt := workflows.EmitWorkflowEvent(scope.Context(), listeners.emitter, topic, payload, listeners.client)
	if receipt.Err == nil {
		return
	}

	listeners.recordEmitFailure(scope, instance, meta, topic.Name(), payload, receipt)
}

// recordActionFailure records a failed action outcome and emits instance failure
func (l *WorkflowListeners) recordActionFailure(scope *observability.Scope, instance *generated.WorkflowInstance, details actionCompletedDetails, obj *workflows.Object) {
	l.recordEvent(scope, instance, enums.WorkflowEventTypeActionCompleted, details.ActionKey, details)
	l.emitInstanceCompleted(scope, instance, enums.WorkflowInstanceStateFailed, obj)
}

// loadActionObject resolves the workflow object for a workflow instance
func (l *WorkflowListeners) loadActionObject(ctx context.Context, instanceID, orgID string) (*workflows.Object, error) {
	objRef, err := loadWorkflowObjectRef(ctx, l.client, instanceID, orgID)
	if err != nil {
		return nil, err
	}

	return workflows.ObjectFromRef(objRef)
}

// evaluateActionWhen resolves the action object and evaluates When expressions
func (l *WorkflowListeners) evaluateActionWhen(ctx context.Context, instance *generated.WorkflowInstance, action models.WorkflowAction, obj *workflows.Object, orgID string) (*workflows.Object, bool, error) {
	if action.When == "" {
		return obj, true, nil
	}

	loadedObj, err := l.loadActionObject(ctx, instance.ID, orgID)
	if err != nil {
		return obj, false, err
	}

	shouldExecute, err := l.engine.EvaluateActionWhen(ctx, action.When, instance, loadedObj)
	if err != nil {
		return loadedObj, false, err
	}

	return loadedObj, shouldExecute, nil
}

// recordEvent persists a workflow event for auditing and debugging
func (l *WorkflowListeners) recordEvent(scope *observability.Scope, instance *generated.WorkflowInstance, eventType enums.WorkflowEventType, actionKey string, details any) {
	if err := persistWorkflowEvent(scope.Context(), l.client, instance, eventType, actionKey, details); err != nil {
		scope.Warn(err, observability.Fields{
			workflowevent.FieldWorkflowInstanceID: instance.ID,
			workflowevent.FieldEventType:          eventType.String(),
		})
	}
}

// recordEmitFailure records an enqueue failure for listener-emitted events
func (l *WorkflowListeners) recordEmitFailure(scope *observability.Scope, instance *generated.WorkflowInstance, meta workflows.EmitFailureMeta, topic string, payload any, receipt workflows.EmitReceipt) {
	details, err := workflows.NewEmitFailureDetails(topic, receipt.EventID, payload, meta, receipt.Err)
	if err != nil {
		scope.Warn(err, observability.Fields{
			workflowevent.FieldWorkflowInstanceID: instance.ID,
			workflowevent.FieldEventType:          enums.WorkflowEventTypeEmitFailed.String(),
		})
		return
	}

	l.recordEvent(scope, instance, enums.WorkflowEventTypeEmitFailed, meta.ActionKey, details)
}

// persistWorkflowEvent stores a workflow event payload for an instance
func persistWorkflowEvent(ctx context.Context, client *generated.Client, instance *generated.WorkflowInstance, eventType enums.WorkflowEventType, actionKey string, details any) error {
	allowCtx := workflows.AllowContext(ctx)

	payload := models.WorkflowEventPayload{
		EventType: eventType,
		ActionKey: actionKey,
	}

	if details != nil {
		encoded, err := json.Marshal(details)
		if err != nil {
			return err
		}
		payload.Details = encoded
	}

	create := client.WorkflowEvent.Create().
		SetWorkflowInstanceID(instance.ID).
		SetEventType(eventType).
		SetPayload(payload)
	ownerID, ownerErr := workflows.ResolveOwnerID(ctx, instance.OwnerID)
	if ownerErr != nil {
		return ownerErr
	}
	create.SetOwnerID(ownerID)

	return create.Exec(allowCtx)
}

// recordEmitFailure records an enqueue failure for engine-emitted events
func recordEmitFailure(ctx context.Context, client *generated.Client, instance *generated.WorkflowInstance, meta workflows.EmitFailureMeta, topic string, payload any, receipt workflows.EmitReceipt) {
	details, err := workflows.NewEmitFailureDetails(topic, receipt.EventID, payload, meta, receipt.Err)
	if err != nil {
		observability.WarnEngine(ctx, observability.OpExecuteAction, meta.EventType.String(), observability.Fields{
			workflowevent.FieldWorkflowInstanceID: instance.ID,
		}, err)
		return
	}

	if err := persistWorkflowEvent(ctx, client, instance, enums.WorkflowEventTypeEmitFailed, meta.ActionKey, details); err != nil {
		observability.WarnEngine(ctx, observability.OpExecuteAction, meta.EventType.String(), observability.Fields{
			workflowevent.FieldWorkflowInstanceID: instance.ID,
		}, err)
	}
}

// recordAssignmentsCreated stores a single assignment-created event for UI history
func (e *WorkflowEngine) recordAssignmentsCreated(ctx context.Context, instance *generated.WorkflowInstance, details assignmentCreatedDetails) {
	if err := persistWorkflowEvent(ctx, e.client, instance, enums.WorkflowEventTypeAssignmentCreated, details.ActionKey, details); err != nil {
		observability.WarnEngine(ctx, observability.OpExecuteAction, enums.WorkflowActionTypeApproval.String(), observability.Fields{
			workflowevent.FieldWorkflowInstanceID: instance.ID,
		}, err)
	}
}
