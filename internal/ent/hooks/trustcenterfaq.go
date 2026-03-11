package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/logx"
)

// HookTrustCenterFAQ sets the trustcenter ID on the note so it is
// always accessible.
func HookTrustCenterFAQ() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterFAQFunc(func(ctx context.Context, m *generated.TrustCenterFAQMutation) (generated.Value, error) {
			id, _ := m.TrustCenterID()
			noteID, _ := m.NoteID()

			v, err := next.Mutate(ctx, m)
			if err != nil {
				return v, err
			}

			if id == "" || noteID == "" {
				return v, nil
			}

			if err := m.Client().Note.UpdateOneID(noteID).
				SetTrustCenterID(id).
				Exec(ctx); err != nil {
				logx.FromContext(ctx).Warn().Err(err).
					Str("note_id", noteID).
					Str("trust_center_id", id).
					Msg("failed to set trust center id on faq note")
			}

			return v, nil
		})
	}, ent.OpCreate)
}
