package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/rs/zerolog"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/usersetting"
	"github.com/theopenlane/core/internal/ent/generated/webauthn"
)

func HookWebauthDelete() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.WebauthnFunc(func(ctx context.Context, m *generated.WebauthnMutation) (generated.Value, error) {
			deletedID, ok := m.ID()
			if !ok {
				return nil, ErrInternalServerError
			}

			passkey, err := m.Client().Webauthn.Get(ctx, deletedID)
			if err != nil {
				zerolog.Ctx(ctx).Error().Err(err).Msg("could not fetch webauthn to delete")

				return nil, err
			}

			count, err := m.Client().Webauthn.Query().
				Where(webauthn.OwnerID(passkey.OwnerID)).
				Count(ctx)
			if err != nil {
				zerolog.Ctx(ctx).Error().Err(err).Msg("could not get count of webauthns")

				return nil, err
			}

			// 1 since this tx is not complete yet
			if count == 1 {
				_, err := m.Client().UserSetting.Update().Where(usersetting.UserID(passkey.OwnerID)).
					SetIsWebauthnAllowed(false).
					Save(ctx)
				if err != nil {
					zerolog.Ctx(ctx).Error().Err(err).Msg("could not disable webauthn from user")

					return nil, err
				}
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpDeleteOne|ent.OpDelete)
}
