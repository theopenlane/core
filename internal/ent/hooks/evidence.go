package hooks

import (
	"context"
	"time"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/objects"
)

// HookEvidenceFiles runs on evidence mutations to check for uploaded files
func HookEvidenceFiles() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.EvidenceFunc(func(ctx context.Context, m *generated.EvidenceMutation) (generated.Value, error) {
			if !isDeleteOp(ctx, m) {
				// validate creation date if only
				// - it is a create operation
				// - it was provided in an update operation
				creationDate, ok := m.CreationDate()
				op := m.Op()

				if op == ent.OpCreate && !ok {
					return nil, ErrZeroTimeNotAllowed
				}

				if ok || op == ent.OpCreate {
					if creationDate.After(time.Now()) {
						return nil, ErrFutureTimeNotAllowed
					}
				}
			}

			// check for uploaded files (e.g. avatar image)
			fileIDs := objects.GetFileIDsFromContext(ctx)
			if len(fileIDs) > 0 {
				var err error

				ctx, err = checkEvidenceFiles(ctx, m)
				if err != nil {
					return nil, err
				}

				m.AddFileIDs(fileIDs...)
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate)
}

// checkEvidenceFiles checks if a evidence files are provided and sets the local file ID(s)
func checkEvidenceFiles[T GenericMutation](ctx context.Context, m T) (context.Context, error) {
	key := "evidenceFiles"

	// get the file from the context, if it exists
	file, _ := objects.FilesFromContextWithKey(ctx, key)

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

			ctx = objects.UpdateFileInContextByKey(ctx, key, file[i])
		}
	}

	return ctx, nil
}
