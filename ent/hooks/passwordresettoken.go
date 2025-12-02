package hooks

import (
	"context"
	"time"

	"entgo.io/ent"

	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/hook"
)

// HookPasswordResetToken runs on reset token mutations and sets expires
func HookPasswordResetToken() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.PasswordResetTokenFunc(func(ctx context.Context, m *generated.PasswordResetTokenMutation) (generated.Value, error) {
			expires, _ := m.TTL()
			if expires.IsZero() {
				m.SetTTL(time.Now().UTC().Add(time.Minute * 15).Truncate(time.Microsecond)) //nolint:mnd
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
}
