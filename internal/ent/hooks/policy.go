package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/stoewer/go-strcase"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/objects"
)

// HookPolicy checks to see if we have an uploaded file.
// If we do, use that as the details of the procedure. and also
// use the name of the file
func HookPolicy() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.InternalPolicyFunc(func(ctx context.Context, m *generated.InternalPolicyMutation) (generated.Value, error) {

			_, exists := m.URL()

			switch exists {
			case true:

				if err := importURLToSchema(m); err != nil {
					return nil, err
				}

			default:

				ctx, err := checkPolicyFile(ctx, m)
				if err != nil {
					return nil, err
				}

				if err := importFileToSchema(ctx, m, m.ObjectManager, "policyFile"); err != nil {
					return nil, err
				}

			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}

func checkPolicyFile[T utils.GenericMutation](ctx context.Context, m T) (context.Context, error) {
	key := "policyFile"

	// get the file from the context, if it exists
	file, _ := objects.FilesFromContextWithKey(ctx, key)

	// return early if no file is provided
	if file == nil {
		return ctx, nil
	}

	// we should only have one file
	if len(file) > 1 {
		return ctx, ErrNotSingularUpload
	}

	// this should always be true, but check just in case
	if file[0].FieldName == key {

		file[0].Parent.ID, _ = m.ID()
		file[0].Parent.Type = strcase.SnakeCase(m.Type())

		ctx = objects.UpdateFileInContextByKey(ctx, key, file[0])
	}

	return ctx, nil
}
