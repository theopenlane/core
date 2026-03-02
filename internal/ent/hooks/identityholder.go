package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	pkgobjects "github.com/theopenlane/core/pkg/objects"
)

// HookIdentityHolderFiles runs on identity holder mutations to check for uploaded files
func HookIdentityHolderFiles() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.IdentityHolderFunc(func(ctx context.Context, m *generated.IdentityHolderMutation) (generated.Value, error) {
			fileIDs := pkgobjects.GetFileIDsFromContext(ctx)
			if len(fileIDs) > 0 {
				var err error

				ctx, err = checkIdentityHolderFiles(ctx, m)
				if err != nil {
					return nil, err
				}
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate)
}

func checkIdentityHolderFiles(ctx context.Context, m *generated.IdentityHolderMutation) (context.Context, error) {
	key := "identityHolderFiles"

	files, _ := pkgobjects.FilesFromContextWithKey(ctx, key)
	if len(files) == 0 {
		return ctx, nil
	}

	fileIDs := make([]string, len(files))
	for i, f := range files {
		fileIDs[i] = f.ID
	}

	m.AddFileIDs(fileIDs...)

	return pkgobjects.ProcessFilesForMutation(ctx, m, key)
}
