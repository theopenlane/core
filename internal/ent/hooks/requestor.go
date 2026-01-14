package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/iam/auth"
)

// HookRequestor sets the requestor_id field on create mutations
func HookRequestor() ent.Hook {
	type RequestorID interface {
		SetRequestorID(string)
	}

	return hook.On(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			obj, ok := m.(RequestorID)
			if !ok {
				return next.Mutate(ctx, m)
			}

			userID, err := auth.GetSubjectIDFromContext(ctx)
			if err != nil {
				return nil, err
			}

			obj.SetRequestorID(userID)

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
}
