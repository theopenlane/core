package hooks

import (
	"context"

	"github.com/theopenlane/core/pkg/objects"
)

// processSingleMutationFile applies a single uploaded file to a mutation and processes it
// yes the function inputs are ugly but this is to keep it generic
func processSingleMutationFile[T any](
	ctx context.Context,
	mutation T,
	key string,
	parentType string,
	errTooMany error,
	setID func(T, string),
	idFunc func(T) (string, bool),
	typeFunc func(T) string,
) (context.Context, error) {
	files, _ := objects.FilesFromContextWithKey(ctx, key)
	if len(files) == 0 {
		return ctx, nil
	}

	if len(files) > 1 {
		return ctx, errTooMany
	}

	setID(mutation, files[0].ID)

	adapter := objects.NewGenericMutationAdapter(mutation, idFunc, typeFunc)

	return objects.ProcessFilesForMutation(ctx, adapter, key, parentType)
}
