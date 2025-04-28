package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/rs/zerolog"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/usersetting"
	"github.com/theopenlane/iam/auth"
)

// HookWebauthDelete runs on passkey delete mutations to ensure
// that we update the user's settings if needed
func HookWebauthDelete() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.WebauthnFunc(func(ctx context.Context, m *generated.WebauthnMutation) (generated.Value, error) {

			userID, err := auth.GetSubjectIDFromContext(ctx)
			if err != nil {
				zerolog.Ctx(ctx).Error().Err(err).Msg("could not fetch authenticated user")

				return nil, err
			}

			count, err := m.Client().Webauthn.Query().
				Count(ctx)
			if err != nil {
				zerolog.Ctx(ctx).Error().Err(err).Msg("could not get count of webauthns")

				return nil, err
			}

			// 1 since this tx is not complete yet
			if count == 1 {
				err = m.Client().UserSetting.Update().Where(usersetting.UserID(userID)).
					SetIsWebauthnAllowed(false).
					Exec(ctx)
				if err != nil {
					zerolog.Ctx(ctx).Error().Err(err).Msg("could not disable webauthn from user")

					return nil, err
				}
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpDeleteOne|ent.OpDelete)
}
