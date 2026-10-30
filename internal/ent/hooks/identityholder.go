package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/entx"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/v2/internal/ent/generated/hook"
	"github.com/theopenlane/logx"
)

// HookIdentityHolderSoftDelete clears identity_holder_id on linked directory accounts
// when an identity holder is soft-deleted, preventing stale foreign key references
func HookIdentityHolderSoftDelete() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.IdentityHolderFunc(func(ctx context.Context, m *generated.IdentityHolderMutation) (generated.Value, error) {
			if !entx.CheckIsSoftDeleteType(ctx, m.Type()) {
				return next.Mutate(ctx, m)
			}

			holderID, ok := m.ID()
			if !ok {
				return next.Mutate(ctx, m)
			}

			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			if clearErr := m.Client().DirectoryAccount.Update().
				Where(directoryaccount.IdentityHolderID(holderID)).
				ClearIdentityHolderID().
				Exec(ctx); clearErr != nil {
				logx.FromContext(ctx).Error().Err(clearErr).Str("identity_holder_id", holderID).Msg("failed to clear identity_holder_id on directory accounts after identity holder soft-delete")

				return retVal, clearErr
			}

			return retVal, nil
		})
	}, ent.OpUpdate|ent.OpUpdateOne)
}
