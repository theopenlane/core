package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
)

// HookTaskCreate runs on task create mutations to set default values that are not provided
// this will set the assigner to the current user if it is not provided
func HookTaskCreate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TaskFunc(func(ctx context.Context, m *generated.TaskMutation) (generated.Value, error) {
			if assigner, _ := m.Assigner(); assigner == "" {
				assigner, err := auth.GetUserIDFromContext(ctx)
				if err != nil {
					return nil, err
				}

				m.SetAssigner(assigner)
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
}
