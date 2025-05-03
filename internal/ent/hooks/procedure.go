package hooks

import (
	"context"
	"fmt"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
)

// HookProcedureSummarize summarizes the procedure and produces a short human readable copy
func HookProcedureSummarize() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.ProcedureFunc(func(ctx context.Context, m *generated.ProcedureMutation) (generated.Value, error) {
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
