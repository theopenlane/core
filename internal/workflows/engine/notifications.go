package engine

import (
	"encoding/json"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/workflowevent"
	"github.com/theopenlane/core/internal/ent/generated/workflowinstance"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/internal/workflows/observability"
)

// notificationExecutionState tracks which conditional notification actions have fired for an instance
type notificationExecutionState struct {
	ExecutedNotifications []string `json:"executed_notifications,omitempty"`
}

// parseNotificationExecutionState reads the execution state from instance context data
func parseNotificationExecutionState(data json.RawMessage) notificationExecutionState {
	if len(data) == 0 {
		return notificationExecutionState{}
	}

	var state notificationExecutionState
	if err := json.Unmarshal(data, &state); err != nil {
		return notificationExecutionState{}
	}

	return state
}

// mergeNotificationExecutionState appends new keys into the existing context data JSON,
// preserving any other keys already stored there
func mergeNotificationExecutionState(data json.RawMessage, newKeys []string) (json.RawMessage, error) {
	var full map[string]any
	if len(data) > 0 {
		if err := json.Unmarshal(data, &full); err != nil {
			full = make(map[string]any)
		}
	} else {
		full = make(map[string]any)
	}

	existing := parseNotificationExecutionState(data)
	existing.ExecutedNotifications = append(existing.ExecutedNotifications, newKeys...)
	full["executed_notifications"] = existing.ExecutedNotifications

	return json.Marshal(full)
}

// getExecutedNotifications returns a set of notification keys already executed for the instance
func (l *WorkflowListeners) getExecutedNotifications(instance *generated.WorkflowInstance) map[string]bool {
	if instance == nil {
		return make(map[string]bool)
	}

	state := parseNotificationExecutionState(instance.Context.Data)
	executed := make(map[string]bool, len(state.ExecutedNotifications))

	for _, key := range state.ExecutedNotifications {
		executed[key] = true
	}

	return executed
}

// trackExecutedNotifications updates the instance context with newly executed notification keys.
// It re-reads the current context from the database before merging to reduce the window
// where concurrent assignment completions could each pass the in-memory dedup guard and
// both drop their key from the persisted state.
func (l *WorkflowListeners) trackExecutedNotifications(scope *observability.Scope, instance *generated.WorkflowInstance, keys []string) {
	ctx := scope.Context()
	allowCtx, orgID, err := workflows.AllowContextWithOrg(ctx)
	if err != nil {
		observability.WarnListener(ctx, observability.OpHandleAssignmentCompleted, "assignment_state_change", observability.Fields{
			workflowevent.FieldWorkflowInstanceID: instance.ID,
		}, err)
		return
	}

	// Re-read the current instance context so the merge is based on the latest committed state
	// rather than the stale in-memory snapshot captured at listener entry.
	current, err := l.client.WorkflowInstance.Query().
		Where(
			workflowinstance.IDEQ(instance.ID),
			workflowinstance.OwnerIDEQ(orgID),
		).
		Only(allowCtx)
	if err != nil {
		observability.WarnListener(ctx, observability.OpHandleAssignmentCompleted, "assignment_state_change", observability.Fields{
			workflowevent.FieldWorkflowInstanceID: instance.ID,
		}, err)
		return
	}

	newData, err := mergeNotificationExecutionState(current.Context.Data, keys)
	if err != nil {
		observability.WarnListener(ctx, observability.OpHandleAssignmentCompleted, "assignment_state_change", observability.Fields{
			workflowevent.FieldWorkflowInstanceID: instance.ID,
		}, err)
		return
	}

	newContext := current.Context
	newContext.Data = newData

	if err := l.client.WorkflowInstance.Update().
		Where(
			workflowinstance.IDEQ(instance.ID),
			workflowinstance.OwnerIDEQ(orgID),
		).
		SetContext(newContext).
		Exec(allowCtx); err != nil {
		observability.WarnListener(ctx, observability.OpHandleAssignmentCompleted, "assignment_state_change", observability.Fields{
			workflowevent.FieldWorkflowInstanceID: instance.ID,
		}, err)
	}
}
