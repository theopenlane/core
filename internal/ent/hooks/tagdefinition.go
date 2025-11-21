package hooks

import (
	"context"
	"strings"

	"entgo.io/ent"
	"github.com/stoewer/go-strcase"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/tagdefinition"
	"github.com/theopenlane/core/pkg/logx"
)

func HookTagDefintion() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TagDefinitionFunc(func(ctx context.Context, m *generated.TagDefinitionMutation) (generated.Value, error) {
			name, ok := m.Name()
			if !ok {
				return next.Mutate(ctx, m)
			}

			ownerID, ok := m.OwnerID()
			if !ok {
				return next.Mutate(ctx, m)
			}

			tagDefs, err := m.Client().TagDefinition.Query().
				Where(tagdefinition.OwnerID(ownerID)).
				All(ctx)
			if err != nil {
				logx.FromContext(ctx).Warn().Err(err).Msg("error querying tag definitions")
				return next.Mutate(ctx, m)
			}

			for _, tagDef := range tagDefs {
				if strings.EqualFold(tagDef.Name, name) {
					return next.Mutate(ctx, m)
				}

				for _, alias := range tagDef.Aliases {
					if strings.EqualFold(alias, name) {
						// found a match! Update the mutation to use the actual tag name
						logx.FromContext(ctx).Debug().
							Str("provided_name", name).
							Str("actual_name", tagDef.Name).
							Msg("resolved alias to tag name")

						m.SetName(tagDef.Name)
						return next.Mutate(ctx, m)
					}
				}
			}

			slug := generateSlug(name)
			m.SetSlug(slug)

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
}

func generateSlug(name string) string { return strcase.LowerCamelCase(name) }
