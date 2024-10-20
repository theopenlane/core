package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
)

func HookEntitlement() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.EntitlementFunc(func(ctx context.Context, m *generated.EntitlementMutation) (generated.Value, error) {
			// set the expires flag if the expires_at field is set
			expiresAt, ok := m.ExpiresAt()
			if ok && !expiresAt.IsZero() {
				m.SetExpires(true)
			}

			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			return retVal, err
		})
	}, ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne)
}
