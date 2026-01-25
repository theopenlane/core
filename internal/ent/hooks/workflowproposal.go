package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
)

// HookWorkflowProposalInvalidateAssignments invalidates approved assignments when a SUBMITTED proposal is edited
func HookWorkflowProposalInvalidateAssignments() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.WorkflowProposalFunc(func(ctx context.Context, m *generated.WorkflowProposalMutation) (generated.Value, error) {
			client := m.Client()
			if !workflowEngineEnabled(client) {
				return next.Mutate(ctx, m)
			}

			// INTENTIONALLY A PLACEHOLDER
			return next.Mutate(ctx, m)

		})
	}, ent.OpUpdate|ent.OpUpdateOne)
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
			if !workflowEngineEnabled(client) {
				return value, nil
			}

			// INTENTIONALLY A PLACEHOLDER

			return value, nil
		})
	}, ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne)
}
