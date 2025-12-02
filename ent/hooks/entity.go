package hooks

import (
	"context"
	"fmt"
	"strings"

	"entgo.io/ent"

	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/hook"
)

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
