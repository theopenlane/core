package hooks

import (
	"context"
	"fmt"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
)

func HookPolicySummarize() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.InternalPolicyFunc(func(ctx context.Context, m *generated.InternalPolicyMutation) (generated.Value, error) {
			details, ok := m.Details()
			if !ok {
				return nil, fmt.Errorf("details does not exists") // nolint:err113
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
