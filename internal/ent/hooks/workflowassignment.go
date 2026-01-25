package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
)

// HookWorkflowAssignmentDecisionAuth ensures only assignment targets can approve/reject.
func HookWorkflowAssignmentDecisionAuth() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.WorkflowAssignmentFunc(func(ctx context.Context, m *generated.WorkflowAssignmentMutation) (generated.Value, error) {
			if !workflowEngineEnabled(m.Client()) {
				return next.Mutate(ctx, m)
			}

			// INTENTIONALLY A PLACEHOLDER

			return next.Mutate(ctx, m)
		})
	}, ent.OpUpdate|ent.OpUpdateOne)
}
