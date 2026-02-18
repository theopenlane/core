package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/workflows"
)

// HookWorkflowInstanceCascadeDelete removes workflow-related child records when instances are deleted.
func HookWorkflowInstanceCascadeDelete() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.WorkflowInstanceFunc(func(ctx context.Context, m *generated.WorkflowInstanceMutation) (generated.Value, error) {
			if !isDeleteOp(ctx, m) {
				return next.Mutate(ctx, m)
			}

			ids := getMutationIDs(ctx, m)
			if len(ids) == 0 {
				return next.Mutate(ctx, m)
			}

			client := m.Client()
			if client == nil {
				return next.Mutate(ctx, m)
			}

			if err := workflows.DeleteWorkflowInstanceChildren(ctx, client, ids); err != nil {
				return nil, err
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpDelete|ent.OpDeleteOne|ent.OpUpdate|ent.OpUpdateOne)
}
