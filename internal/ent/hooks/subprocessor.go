package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	pkgobjects "github.com/theopenlane/core/pkg/objects"
)

func HookSubprocessor() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.SubprocessorFunc(func(ctx context.Context, m *generated.SubprocessorMutation) (generated.Value, error) {
			// check for uploaded files (e.g. logo image)
			fileIDs := pkgobjects.GetFileIDsFromContext(ctx)
			if len(fileIDs) > 0 {
				var err error

				ctx, err = checkSubprocessorLogoFile(ctx, m)
				if err != nil {
					return nil, err
				}

				m.AddFileIDs(fileIDs...)
			}
			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}

func checkSubprocessorLogoFile(ctx context.Context, m *generated.SubprocessorMutation) (context.Context, error) {
	logoKey := "logoFile"

	// get the file from the context, if it exists
	logoFile, _ := pkgobjects.FilesFromContextWithKey(ctx, logoKey)
	if logoFile == nil {
		return ctx, nil
	}

	// this should always be true, but check just in case
	if logoFile[0].FieldName == logoKey {
		// we should only have one file
		if len(logoFile) > 1 {
			return ctx, ErrTooManyLogoFiles
		}
		m.SetLogoLocalFileID(logoFile[0].ID)

		logoFile[0].Parent.ID, _ = m.ID()
		logoFile[0].Parent.Type = "trust_center_setting"

		ctx = pkgobjects.UpdateFileInContextByKey(ctx, logoKey, logoFile[0])
	}

	return ctx, nil
}
