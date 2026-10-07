package hooks

import (
	"context"
	"encoding/json"
	"fmt"

	"entgo.io/ent"
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/hook"
	"github.com/theopenlane/core/v2/internal/workflows"
	"github.com/theopenlane/core/v2/internal/workflows/engine"
	"github.com/theopenlane/core/v2/pkg/urlx"
)

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

// HookWorkflowDefinitionWebhookAddress rejects definitions whose webhook actions target localhost or a
// non-public IP literal. It runs regardless of engine state so definitions saved while the engine is
// disabled are still checked; hostnames are enforced against resolved addresses when the webhook is sent
func HookWorkflowDefinitionWebhookAddress() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.WorkflowDefinitionFunc(func(ctx context.Context, m *generated.WorkflowDefinitionMutation) (generated.Value, error) {
			doc, ok := m.DefinitionJSON()
			if !ok || webhookPrivateAddressesAllowed() {
				return next.Mutate(ctx, m)
			}

			for _, action := range doc.Actions {
				if err := requirePublicWebhookURL(action); err != nil {
					return nil, fmt.Errorf("%w: action %q: %w", ErrWorkflowWebhookURLNotAllowed, action.Key, err)
				}
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne)
}

// webhookPrivateAddressesAllowed reports whether the registered workflow engine permits non-public webhook targets
func webhookPrivateAddressesAllowed() bool {
	wfEngine := engine.Default()

	return wfEngine != nil && wfEngine.WebhookAllowPrivateAddresses()
}

// requirePublicWebhookURL validates the url host of a webhook action, ignoring all other action types
func requirePublicWebhookURL(action models.WorkflowAction) error {
	if lo.FromPtr(enums.ToWorkflowActionType(action.Type)) != enums.WorkflowActionTypeWebhook || len(action.Params) == 0 {
		return nil
	}

	var params workflows.WebhookActionParams
	if err := json.Unmarshal(action.Params, &params); err != nil {
		return err
	}

	target, err := urlx.ParseAbsolute(params.URL)
	if err != nil {
		return err
	}

	return urlx.RequirePublicHost(target)
}
