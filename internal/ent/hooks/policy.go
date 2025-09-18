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

				var err error
				ctx, err = checkPolicyFile(ctx, m)
				if err != nil {
					return nil, err
				}

			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}

func checkPolicyFile[T utils.GenericMutation](ctx context.Context, m T) (context.Context, error) {
	key := "policyFile"

	// Get files using the new helper
	files, err := objects.GetFilesForKey(ctx, key)
	if err != nil {
		return ctx, err
	}

	// Return early if no files
	if len(files) == 0 {
		return ctx, nil
	}

	// we should only have one file
	if len(files) > 1 {
		return ctx, ErrNotSingularUpload
	}

	// Create adapter for the existing mutation interface
	adapter := objects.NewGenericMutationAdapter(m,
		func(mut T) (string, bool) { return mut.ID() },
		func(mut T) string { return mut.Type() },
	)

	// Use the generic helper to process the file with snake_case parent type
	parentType := strcase.SnakeCase(m.Type())
	return objects.ProcessFilesForMutation(ctx, adapter, key, parentType)
}
