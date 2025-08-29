package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/objects"
)

func HookProcedure() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.ProcedureFunc(func(ctx context.Context, m *generated.ProcedureMutation) (generated.Value, error) {

			m.SetName("ues")
			fileIDs := objects.GetFileIDsFromContext(ctx)
			if len(fileIDs) > 0 {
				var err error

				ctx, err = checkProcedureFile(ctx, m)
				if err != nil {
					return nil, err
				}

				m.AddFileIDs(fileIDs...)
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
}

func checkProcedureFile[T utils.GenericMutation](ctx context.Context, m T) (context.Context, error) {
	key := "procedureFile"

	// get the file from the context, if it exists
	file, _ := objects.FilesFromContextWithKey(ctx, key)

	// return early if no file is provided
	if file == nil {
		return ctx, nil
	}

	// we should only have one file
	if len(file) > 1 {
		return ctx, ErrTooManyAvatarFiles
	}

	// this should always be true, but check just in case
	if file[0].FieldName == key {

		file[0].Parent.ID, _ = m.ID()
		file[0].Parent.Type = m.Type()

		ctx = objects.UpdateFileInContextByKey(ctx, key, file[0])
	}

	return ctx, nil
}
