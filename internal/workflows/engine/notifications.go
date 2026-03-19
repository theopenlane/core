package engine

import (
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/workflowevent"
	"github.com/theopenlane/core/internal/ent/generated/workflowinstance"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/internal/workflows/observability"
)

// getExecutedNotifications returns a set of notification keys already executed for the instance
func (l *WorkflowListeners) getExecutedNotifications(instance *generated.WorkflowInstance) map[string]bool {
	if instance == nil {
		return make(map[string]bool)
	}

	keys := instance.Context.ExecutedNotifications
	executed := make(map[string]bool, len(keys))
	for _, key := range keys {
		executed[key] = true
	}

	return executed
}

// trackExecutedNotifications appends newly executed notification keys to the typed field on the
// instance context. It re-reads the current context from the database before merging to reduce
// the race window where concurrent assignment completions could each bypass the in-memory guard.
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

	newContext := current.Context
	newContext.ExecutedNotifications = append(newContext.ExecutedNotifications, keys...)

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
