package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/logx"
)

// HookDirectoryAccountDelete syncs identity holder email aliases after a directory
// account is removed, since the async listener cannot look up the deleted row
func HookDirectoryAccountDelete() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.DirectoryAccountFunc(func(ctx context.Context, m *generated.DirectoryAccountMutation) (generated.Value, error) {
			accountID, ok := m.ID()
			if !ok {
				return next.Mutate(ctx, m)
			}

			// OldIdentityHolderID only works on UpdateOne, not DeleteOne; read the row directly
			account, err := m.Client().DirectoryAccount.Get(ctx, accountID)
			if err != nil || account.IdentityHolderID == nil || *account.IdentityHolderID == "" {
				return next.Mutate(ctx, m)
			}

			capturedHolderID := *account.IdentityHolderID

			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			holder, holderErr := m.Client().IdentityHolder.Get(ctx, capturedHolderID)
			if holderErr != nil {
				logx.FromContext(ctx).Error().Err(holderErr).Str("identity_holder_id", capturedHolderID).Str("directory_account_id", accountID).Msg("failed to load identity holder for alias sync after directory account delete")

				return retVal, holderErr
			}

			if syncErr := syncEmailAliases(ctx, m.Client(), holder); syncErr != nil {
				logx.FromContext(ctx).Error().Err(syncErr).Str("identity_holder_id", capturedHolderID).Str("directory_account_id", accountID).Msg("failed to sync email aliases after directory account delete")

				return retVal, syncErr
			}

			return retVal, nil
		})
	}, ent.OpDeleteOne)
}
