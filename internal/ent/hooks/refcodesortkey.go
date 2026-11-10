package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/v2/internal/ent/generated/hook"
	"github.com/theopenlane/core/v2/pkg/naturalsort"
)

// refCodeSortKeyMutation is a mutation of a schema whose ref_code is ordered through ref_code_sort_key
type refCodeSortKeyMutation interface {
	RefCode() (string, bool)
	SetRefCodeSortKey(string)
}

// HookRefCodeSortKey keeps ref_code_sort_key in step with ref_code on every write that sets ref_code
func HookRefCodeSortKey() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if mut, ok := m.(refCodeSortKeyMutation); ok {
				if refCode, set := mut.RefCode(); set {
					mut.SetRefCodeSortKey(naturalsort.Key(refCode))
				}
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne)
}
