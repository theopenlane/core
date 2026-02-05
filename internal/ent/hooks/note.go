package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/discussion"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/note"
	pkgobjects "github.com/theopenlane/core/pkg/objects"
)

// HookNoteFiles runs on note mutations to check for uploaded files
func HookNoteFiles() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.NoteFunc(func(ctx context.Context, m *generated.NoteMutation) (generated.Value, error) {
			fileIDs := pkgobjects.GetFileIDsFromContext(ctx)
			if len(fileIDs) > 0 {
				var err error

				ctx, err = checkNoteFiles(ctx, m)
				if err != nil {
					return nil, err
				}
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate)
}

// HookDeleteDiscussionOnLastComment deletes the discussion when the last comment is deleted
func HookDeleteDiscussionOnLastComment() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.NoteFunc(func(ctx context.Context, m *generated.NoteMutation) (generated.Value, error) {
			if !isDeleteOp(ctx, m) {
				return next.Mutate(ctx, m)
			}

			id, ok := m.ID()
			if !ok {
				return next.Mutate(ctx, m)
			}

			obj, err := m.Client().Note.Get(ctx, id)
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

			if err := m.Client().Discussion.DeleteOneID(discussionID).Exec(ctx); err != nil {
				return v, err
			}

			return v, nil
		})
	}, ent.OpDeleteOne|ent.OpUpdateOne)
}

// checkNoteFiles checks if note files are provided and sets the local file ID(s)
func checkNoteFiles(ctx context.Context, m *generated.NoteMutation) (context.Context, error) {
	key := "noteFiles"

	files, _ := pkgobjects.FilesFromContextWithKey(ctx, key)
	if len(files) == 0 {
		return ctx, nil
	}

	fileIDs := make([]string, len(files))
	for i, f := range files {
		fileIDs[i] = f.ID
	}

	m.AddFileIDs(fileIDs...)

	adapter := pkgobjects.NewGenericMutationAdapter(m,
		func(mut *generated.NoteMutation) (string, bool) { return mut.ID() },
		func(mut *generated.NoteMutation) string { return mut.Type() },
	)

	return pkgobjects.ProcessFilesForMutation(ctx, adapter, key)
}
