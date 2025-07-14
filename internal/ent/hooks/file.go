package hooks

import (
	"context"
	"errors"

	"entgo.io/ent"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
)

var (
	errInvalidStoragePath = errors.New("invalid path when deleting file from object storage")
)

// HookFileDelete makes sure to clean up the file from external storage once deleted
func HookFileDelete() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.FileFunc(
			func(ctx context.Context, m *generated.FileMutation) (generated.Value, error) {

				var storagePath string
				if m.ObjectManager != nil && isDeleteOp(ctx, m) {

					id, ok := m.ID()
					if !ok {
						return nil, errInvalidStoragePath
					}

					file, err := m.Client().File.Get(ctx, id)
					if err != nil {
						return nil, err
					}

					storagePath = file.StoragePath
				}

				v, err := next.Mutate(ctx, m)
				if err != nil {
					return nil, err
				}

				if storagePath != "" {
					if err := m.ObjectManager.Storage.Delete(ctx, storagePath); err != nil {
						return nil, err
					}
				}

				return v, err
			})
	}, ent.OpDelete|ent.OpDeleteOne|ent.OpUpdate|ent.OpUpdateOne)
}
