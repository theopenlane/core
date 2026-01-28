package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/workflows"
)

// HookWorkflowDefinitionPrefilter derives prefilter fields from the definition JSON.
func HookWorkflowDefinitionPrefilter() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.WorkflowDefinitionFunc(func(ctx context.Context, m *generated.WorkflowDefinitionMutation) (generated.Value, error) {
			if !workflowEngineEnabled(ctx, m.Client()) {
				return next.Mutate(ctx, m)
			}

			doc, ok := m.DefinitionJSON()
			if !ok {
				return next.Mutate(ctx, m)
			}

			operations, fields := workflows.DeriveTriggerPrefilter(doc)
			if len(operations) == 0 {
				m.SetTriggerOperations(nil)
			} else {
				m.SetTriggerOperations(operations)
			}

			if len(fields) == 0 {
				m.SetTriggerFields(nil)
			} else {
				m.SetTriggerFields(fields)
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne)
}
