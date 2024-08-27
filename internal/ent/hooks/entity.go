package hooks

import (
	"context"
	"fmt"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/utils/ulids"
)

// HookEntityCreate runs on entity mutations to set default values that are not provided
func HookEntityCreate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.EntityFunc(func(ctx context.Context, mutation *generated.EntityMutation) (generated.Value, error) {
			// require either a display name or a name
			displayName, _ := mutation.DisplayName()
			name, _ := mutation.Name()

			// exit early if we have no name
			if displayName == "" && name == "" {
				return nil, ErrMissingRequiredName
			}

			// set display name based on name if it isn't set
			if displayName == "" {
				mutation.SetDisplayName(name)
			}

			// set unique name based on display name if it isn't set
			if name == "" {
				uniqueName := fmt.Sprintf("%s-%s", displayName, ulids.New().String())
				mutation.SetName(uniqueName)
			}

			return next.Mutate(ctx, mutation)
		})
	}, ent.OpCreate)
}
