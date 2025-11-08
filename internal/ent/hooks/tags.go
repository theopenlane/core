package hooks

import (
	"context"
	"slices"

	"entgo.io/ent"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/tagdefinition"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
)

// tagMutation is an interface for mutations that have tags
type tagMutation interface {
	utils.GenericMutation

	Tags() ([]string, bool)
	AppendedTags() ([]string, bool)
}

// HookTags will create tag definitions if they do not already exist when tags are added to an entity
func HookTags() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			mut := m.(tagMutation)

			tags, ok := mut.Tags()
			appendTags, appendOk := mut.AppendedTags()

			if !ok && !appendOk {
				return next.Mutate(ctx, m)
			}

			// get all tags to be added
			newTags := slices.Concat(tags, appendTags)
			uniqueTags := lo.Uniq(newTags)

			// for each tag, create the tag definition if it does not already exist
			for _, tag := range uniqueTags {
				if tag == "" {
					continue
				}

				exists, err := mut.Client().TagDefinition.Query().
					Where(tagdefinition.NameEqualFold(tag)).
					Exist(ctx)
				if !exists {
					input := generated.CreateTagDefinitionInput{
						Name: tag,
					}

					if err := mut.Client().TagDefinition.Create().
						SetInput(input).
						Exec(ctx); err != nil {
						log.Warn().Err(err).Str("tag", tag).Msg("error creating tag definition")
					}
				} else if err != nil {
					log.Warn().Err(err).Msg("error querying tag definitions, skipping org tag creation")
				}
			}

			// continue with the rest of the mutation
			return next.Mutate(ctx, m)
		})
	}, hook.HasOp(ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne))
}
