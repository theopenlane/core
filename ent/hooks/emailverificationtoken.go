package hooks

import (
	"context"
	"time"

	"entgo.io/ent"

	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/hook"
)

// HookEmailVerificationToken runs on email verification mutations and sets expires
func HookEmailVerificationToken() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.EmailVerificationTokenFunc(func(ctx context.Context, m *generated.EmailVerificationTokenMutation) (generated.Value, error) {
			expires, _ := m.TTL()
			if expires.IsZero() {
				m.SetTTL(time.Now().UTC().Add(time.Hour * 24 * 7)) //nolint:mnd
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
}
