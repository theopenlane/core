package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
)

// HookAssetCreate sets the display name for assets everytime one is created
func HookAssetCreate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.AssetFunc(func(ctx context.Context, m *generated.AssetMutation) (generated.Value, error) {
			name, _ := m.DisplayName()
			if name == "" {
				name, _ := m.Name()
				m.SetDisplayName(name)
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
}
