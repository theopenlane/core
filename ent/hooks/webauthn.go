package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/hook"
	"github.com/theopenlane/shared/logx"
)

// HookWebauthnDelete runs on passkey delete mutations to ensure
// that we update the user's settings if needed
func HookWebauthnDelete() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.WebauthnFunc(func(ctx context.Context, m *generated.WebauthnMutation) (generated.Value, error) {
			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			// get the count of webauthns for the user
			count, err := m.Client().Webauthn.Query().
				Count(ctx)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("could not get count of webauthns")

				return nil, err
			}

			// if the count is 0, we need to disable webauthn for the user
			if count == 0 {
				if err = m.Client().UserSetting.Update().
					SetIsWebauthnAllowed(false).
					Exec(ctx); err != nil {
					logx.FromContext(ctx).Error().Err(err).Msg("could not disable webauthn from user")

					return nil, err
				}
			}

			return retVal, nil
		})
	}, ent.OpDeleteOne|ent.OpDelete)
}
