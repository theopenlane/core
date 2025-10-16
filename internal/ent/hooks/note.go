package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
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

				m.AddFileIDs(fileIDs...)
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate)
}

// checkNoteFiles checks if note files are provided and sets the local file ID(s)
func checkNoteFiles[T utils.GenericMutation](ctx context.Context, m T) (context.Context, error) {
	key := "noteFiles"

	files, _ := pkgobjects.FilesFromContextWithKey(ctx, key)
	if len(files) == 0 {
		return ctx, nil
	}

	adapter := pkgobjects.NewGenericMutationAdapter(m,
		func(mut T) (string, bool) { return mut.ID() },
		func(mut T) string { return mut.Type() },
	)

	return pkgobjects.ProcessFilesForMutation(ctx, adapter, key)
}
