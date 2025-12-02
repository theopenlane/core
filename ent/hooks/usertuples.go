package hooks

import (
	"context"

	"entgo.io/ent"
)

// HookUserCanViewTuple adds the user#can_view relation for the created object
// it is agnostic to the object type so it can be used on any schema
func HookUserCanViewTuple() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			if err := addUserCanViewRelation(ctx, m); err != nil {
				return nil, err
			}

			return retVal, nil
		})
	}
}
