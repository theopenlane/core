package hooks

import (
	"context"
	"strings"

	"entgo.io/ent"

	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/hook"
)

// HookContact runs on contact create mutations
func HookContact() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return hook.ContactFunc(func(ctx context.Context, m *generated.ContactMutation) (generated.Value, error) {
			email, ok := m.Email()
			if ok {
				// lowercase the email for uniqueness
				m.SetEmail(strings.ToLower(email))
			}

			return next.Mutate(ctx, m)
		})
	},
		hook.And(
			hook.HasFields("email"),
			hook.HasOp(ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate),
		),
	)
}
