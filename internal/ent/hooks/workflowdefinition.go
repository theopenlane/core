package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/hook"
	"github.com/theopenlane/core/v2/internal/workflows"
	"github.com/theopenlane/core/v2/internal/workflows/engine"
)

// ValidateWorkflowDefinitionWebhooks is the definition_json field validator; it honors the registered
// engine's private address setting and stays strict when no engine is registered
func ValidateWorkflowDefinitionWebhooks(doc models.WorkflowDefinitionDocument) error {
	wfEngine := engine.Default()

	return workflows.ValidateWebhookDestinations(doc, wfEngine != nil && wfEngine.WebhookAllowPrivateAddresses())
}

// HookWorkflowDefinitionPrefilter derives prefilter fields from the definition JSON.
func HookWorkflowDefinitionPrefilter() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.WorkflowDefinitionFunc(func(ctx context.Context, m *generated.WorkflowDefinitionMutation) (generated.Value, error) {
			if !workflowEngineEnabled() {
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
