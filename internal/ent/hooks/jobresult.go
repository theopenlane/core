package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	pkgobjects "github.com/theopenlane/core/pkg/objects"
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

	// get the file from the context, if it exists
	file, _ := pkgobjects.FilesFromContextWithKey(ctx, key)

	// return early if no file is provided
	if file == nil {
		return ctx, nil
	}

	// set the parent ID and type for the file(s)
	for i, f := range file {
		// this should always be true, but check just in case
		if f.FieldName == key {
			file[i].Parent.ID, _ = m.ID()
			file[i].Parent.Type = m.Type()

			ctx = pkgobjects.UpdateFileInContextByKey(ctx, key, file[i])
		}
	}

	return ctx, nil
}
