package engine

import (
	"encoding/json"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/workflowevent"
	"github.com/theopenlane/core/internal/ent/generated/workflowinstance"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/internal/workflows/observability"
)

// getExecutedNotifications returns a set of notification keys already executed for the instance.
func (l *WorkflowListeners) getExecutedNotifications(instance *generated.WorkflowInstance) map[string]bool {
	executed := make(map[string]bool)
	if instance == nil || len(instance.Context.Data) == 0 {
		return executed
	}

	var data map[string]any
	if err := json.Unmarshal(instance.Context.Data, &data); err != nil {
		return executed
	}

	raw, ok := data["executed_notifications"]
	if !ok || raw == nil {
		return executed
	}

	switch values := raw.(type) {
	case []any:
		for _, v := range values {
			if s, ok := v.(string); ok {
				executed[s] = true
			}
		}
	case []string:
		for _, s := range values {
			executed[s] = true
		}
	}

	return executed
}

// trackExecutedNotifications updates the instance context with newly executed notification keys
func (l *WorkflowListeners) trackExecutedNotifications(scope *observability.Scope, instanceID string, keys []string) {
	ctx := scope.Context()
	allowCtx, orgID, err := workflows.AllowContextWithOrg(ctx)
	if err != nil {
		observability.WarnListener(ctx, observability.OpHandleAssignmentCompleted, "assignment_state_change", observability.Fields{
			workflowevent.FieldWorkflowInstanceID: instanceID,
		}, err)
		return
	}

	instance, err := l.client.WorkflowInstance.Query().
		Where(
			workflowinstance.IDEQ(instanceID),
			workflowinstance.OwnerIDEQ(orgID),
		).
		Only(allowCtx)
	if err != nil {
		observability.WarnListener(ctx, observability.OpHandleAssignmentCompleted, "assignment_state_change", observability.Fields{
			workflowevent.FieldWorkflowInstanceID: instanceID,
		}, err)
		return
	}

	// Parse existing data
	var data map[string]any
	if len(instance.Context.Data) > 0 {
		if err := json.Unmarshal(instance.Context.Data, &data); err != nil {
			data = make(map[string]any)
		}
	} else {
		data = make(map[string]any)
	}

	// Get existing executed notifications
	var existing []string
	if executedRaw, ok := data["executed_notifications"].([]any); ok {
		for _, v := range executedRaw {
			if s, ok := v.(string); ok {
				existing = append(existing, s)
			}
		}
	}

	// Add new keys
	existing = append(existing, keys...)
	data["executed_notifications"] = existing

	// Marshal back to json.RawMessage
	newData, err := json.Marshal(data)
	if err != nil {
		observability.WarnListener(ctx, observability.OpHandleAssignmentCompleted, "assignment_state_change", observability.Fields{
			workflowevent.FieldWorkflowInstanceID: instanceID,
		}, err)
		return
	}

	newContext := instance.Context
	newContext.Data = newData

	if err := l.client.WorkflowInstance.Update().
		Where(
			workflowinstance.IDEQ(instanceID),
			workflowinstance.OwnerIDEQ(orgID),
		).
		SetContext(newContext).
		Exec(allowCtx); err != nil {
		observability.WarnListener(ctx, observability.OpHandleAssignmentCompleted, "assignment_state_change", observability.Fields{
			workflowevent.FieldWorkflowInstanceID: instanceID,
		}, err)
	}
}
