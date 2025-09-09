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

				if m.ObjectManager == nil && !isDeleteOp(ctx, m) {
					return next.Mutate(ctx, m)
				}

				var ids []string

				switch m.Op() {
				case ent.OpDelete:
					dbIDs, err := m.IDs(ctx)
					if err != nil {
						return nil, err
					}

					ids = append(ids, dbIDs...)

				case ent.OpDeleteOne:

					id, ok := m.ID()
					if !ok {
						return nil, errInvalidStoragePath
					}

					ids = append(ids, id)
				}

				v, err := next.Mutate(ctx, m)
				if err != nil {
					return nil, err
				}

				for _, id := range ids {

					file, err := m.Client().File.Get(ctx, id)
					if err != nil {
						return nil, err
					}

					if file.StoragePath != "" {
						if err := m.ObjectManager.Storage.Delete(ctx, file.StoragePath); err != nil {
							return nil, err
						}
					}
				}

				return v, err
			})
	}, ent.OpDelete|ent.OpDeleteOne|ent.OpUpdate|ent.OpUpdateOne)
}
