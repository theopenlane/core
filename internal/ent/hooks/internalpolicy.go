package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
)

// HookPolicySummarize summarizes the policy and produces a short human readable copy
func HookPolicySummarize() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.InternalPolicyFunc(func(ctx context.Context, m *generated.InternalPolicyMutation) (generated.Value, error) {
			details, ok := m.Details()
			if !ok {
				return next.Mutate(ctx, m)
			}

			summarized, err := m.Summarizer.Summarize(ctx, details)
			if err != nil {
				return nil, err
			}

			m.SetSummary(summarized)

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne)
}
