package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated/hook"
)

// HookImpersonatorAttribution stamps the acting impersonator onto a record when the mutation is
// performed during an impersonation session (for example an Openlane support session). The record's
// created_by/updated_by still reflect the impersonated identity, while updated_by_impersonator
// records the real actor behind the session so changes remain traceable to a specific person. When
// the mutation is not impersonated the field is cleared so it reflects the most recent actor
func HookImpersonatorAttribution() ent.Hook {
	type impersonatorAttribution interface {
		SetUpdatedByImpersonator(string)
		ClearUpdatedByImpersonator()
	}

	return hook.On(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			obj, ok := m.(impersonatorAttribution)
			if !ok {
				return next.Mutate(ctx, m)
			}

			if caller, hasCaller := auth.CallerFromContext(ctx); hasCaller && caller != nil && caller.IsImpersonated() {
				obj.SetUpdatedByImpersonator(caller.Impersonation.ImpersonatorID)
			} else {
				obj.ClearUpdatedByImpersonator()
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne)
}
