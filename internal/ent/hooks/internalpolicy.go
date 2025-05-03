package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
)

func HookPolicySummarize() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.ProcedureFunc(func(ctx context.Context, m *generated.ProcedureMutation) (generated.Value, error) {

			retValue, err := next.Mutate(ctx, m)

			return retValue, err
		})
	}, ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne)
}
