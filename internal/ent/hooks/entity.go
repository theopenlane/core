package hooks

import (
	"context"
	"fmt"
	"strings"

	"entgo.io/ent"

	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	pkgobjects "github.com/theopenlane/core/pkg/objects"
)

// HookEntityFiles runs on entity mutations to check for uploaded files
func HookEntityFiles() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.EntityFunc(func(ctx context.Context, m *generated.EntityMutation) (generated.Value, error) {
			fileIDs := pkgobjects.GetFileIDsFromContext(ctx)
			if len(fileIDs) > 0 {
				var err error

				ctx, err = checkEntityFiles(ctx, m)
				if err != nil {
					return nil, err
				}
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate)
}

func checkEntityFiles(ctx context.Context, m *generated.EntityMutation) (context.Context, error) {
	key := "entityFiles"

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

// HookEntityCreate runs on entity mutations to set default values that are not provided
func HookEntityCreate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.EntityFunc(func(ctx context.Context, m *generated.EntityMutation) (generated.Value, error) {
			// require either a display name or a name
			displayName, _ := m.DisplayName()
			name, _ := m.Name()

			// exit early if we have no name
			if displayName == "" && name == "" {
				return nil, ErrMissingRequiredName
			}

			// set display name based on name if it isn't set
			if displayName == "" {
				m.SetDisplayName(name)
			}

			// set unique name based on display name if it isn't set
			if name == "" {
				uniqueName := fmt.Sprintf("%s-%s", displayName, ulids.New().String())
				m.SetName(uniqueName)
			}

			// trim trailing whitespace from the name
			name, _ = m.Name() // re-fetch incase it was updated above
			m.SetName(strings.TrimSpace(name))

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
}
