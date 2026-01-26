package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
)

// HookWorkflowApprovalRouting intercepts mutations on workflowable schemas and routes them
// to WorkflowProposal when a matching workflow definition with approval requirements exists.
// This enables the "proposed changes" pattern where mutations require approval before being applied.
func HookWorkflowApprovalRouting() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			mut, ok := m.(utils.GenericMutation)
			if !ok {
				return next.Mutate(ctx, m)
			}

			client := mut.Client()
			if !workflowEngineEnabled(client) {
				return next.Mutate(ctx, m)
			}

			// INTENTIONALLY A PLACEHOLDER

			// Route to proposed changes instead of applying directly
			return nil, nil
		})
	}, ent.OpUpdate|ent.OpUpdateOne)
}
