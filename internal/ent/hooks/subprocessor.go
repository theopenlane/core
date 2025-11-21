package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	pkgobjects "github.com/theopenlane/core/pkg/objects"
)

// HookSubprocessor runs on subprocessor mutations to check for uploaded logo file
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
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}

func checkSubprocessorLogoFile(ctx context.Context, m *generated.SubprocessorMutation) (context.Context, error) {
	logoKey := "logoFile"

	logoFiles, _ := pkgobjects.FilesFromContextWithKey(ctx, logoKey)
	if len(logoFiles) == 0 {
		return ctx, nil
	}

	if len(logoFiles) > 1 {
		return ctx, ErrTooManyLogoFiles
	}

	m.SetLogoFileID(logoFiles[0].ID)

	adapter := pkgobjects.NewGenericMutationAdapter(m,
		func(mut *generated.SubprocessorMutation) (string, bool) { return mut.ID() },
		func(mut *generated.SubprocessorMutation) string { return mut.Type() },
	)

	return pkgobjects.ProcessFilesForMutation(ctx, adapter, logoKey)
}
