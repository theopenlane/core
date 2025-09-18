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

	// get the file from the context, if it exists
	file, err := pkgobjects.FilesFromContextWithKey(ctx, key)
	if err != nil {
		return ctx, err
	}

	if file == nil {
		return ctx, nil
	}

	for i, f := range file {
		if f.FieldName == key {
			file[i].Parent.ID, _ = m.ID()
			file[i].Parent.Type = m.Type()

			ctx = pkgobjects.UpdateFileInContextByKey(ctx, key, file[i])
		}
	}

	return ctx, nil
}
