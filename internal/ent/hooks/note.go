package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/discussion"
	"github.com/theopenlane/core/v2/internal/ent/generated/hook"
	"github.com/theopenlane/core/v2/internal/ent/generated/note"
	"github.com/theopenlane/iam/auth"
)

// HookDeleteDiscussionDelete deletes the discussion when the last comment is deleted
func HookDeleteDiscussion() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.NoteFunc(func(ctx context.Context, m *generated.NoteMutation) (generated.Value, error) {
			if !isDeleteOp(ctx, m) {
				return next.Mutate(ctx, m)
			}

			id, ok := m.ID()
			if !ok {
				return next.Mutate(ctx, m)
			}

			obj, err := m.Client().Note.Query().
				Select(note.FieldDiscussionID).
				Where(note.ID(id)).
				Only(ctx)
			if err != nil {
				return nil, err
			}

			if obj.DiscussionID == "" {
				return next.Mutate(ctx, m)
			}

			discussionID := obj.DiscussionID

			v, err := next.Mutate(ctx, m)
			if err != nil {
				return v, err
			}

			count, err := m.Client().Note.Query().
				Where(note.HasDiscussionWith(discussion.ID(discussionID))).
				Count(ctx)
			if err != nil {
				return v, err
			}

			if count > 0 {
				return v, nil
			}

			internalCtx := auth.WithInternalOperationContext(ctx)

			if err := m.Client().Discussion.DeleteOneID(discussionID).Exec(internalCtx); err != nil {
				return v, err
			}

			return v, nil
		})
	}, ent.OpDeleteOne|ent.OpDelete|ent.OpUpdateOne|ent.OpUpdate)
}
