package hooks

import (
	"context"

	"entgo.io/ent"

	pkgobjects "github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/hook"
	"github.com/theopenlane/ent/privacy/utils"
)

// HookJobResultFiles runs on jobresult mutations to check for uploaded files
func HookJobResultFiles() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.JobResultFunc(func(ctx context.Context, m *generated.JobResultMutation) (generated.Value, error) {
			// check for uploaded files
			fileIDs := pkgobjects.GetFileIDsFromContext(ctx)
			if len(fileIDs) > 0 {
				var err error

				ctx, err = checkJobResultFiles(ctx, m)
				if err != nil {
					return nil, err
				}

				// JobResult has a file relationship, so we set the file ID
				if len(fileIDs) > 0 {
					m.SetFileID(fileIDs[0]) // JobResult has single file relationship
				}
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate)
}

// checkJobResultFiles checks if jobresult files are provided and sets the local file ID(s)
func checkJobResultFiles[T utils.GenericMutation](ctx context.Context, m T) (context.Context, error) {
	key := "jobResultFiles"

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
