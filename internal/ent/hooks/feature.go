package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
)

func HookFeature() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.FeatureFunc(func(ctx context.Context, m *generated.FeatureMutation) (generated.Value, error) {
			// set the display name if it is not set
			if m.Op() == ent.OpCreate {
				displayName, _ := m.DisplayName()
				if displayName == "" {
					name, ok := m.Name()
					if ok {
						m.SetDisplayName(name)
					}
				}
			}

			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			return retVal, err
		})
	}, ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne)
}
