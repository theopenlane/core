package hooks

import (
	"context"
	"slices"
	"strings"

	"entgo.io/ent"

	"github.com/samber/lo"
	"github.com/stoewer/go-strcase"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/hook"
	"github.com/theopenlane/core/v2/internal/ent/generated/tagdefinition"
	"github.com/theopenlane/core/v2/internal/ent/privacy/utils"
	"github.com/theopenlane/core/v2/pkg/logx"
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

			// only auto-create tags when we have an organization ID in context
			// this ensures we do not create global tags automatically from internal requests without
			// organization context
			orgID, err := auth.GetOrganizationIDFromContext(ctx)
			if err != nil {
				return next.Mutate(ctx, m)
			}

			// for each tag, create the tag definition if it does not already exist
			for _, tag := range uniqueTags {
				if tag == "" {
					continue
				}

				// match on slug as well as name, the unique index is on both and kebab casing collapses
				// names that differ only in spacing or punctuation
				exists, err := mut.Client().TagDefinition.Query().
					Where(
						tagdefinition.Or(
							tagdefinition.NameEqualFold(tag),
							tagdefinition.SlugEqualFold(strcase.KebabCase(strings.TrimSpace(tag))),
						),
					).
					Exist(ctx)
				if err != nil {
					logx.FromContext(ctx).Error().Err(err).Str("tag", tag).Msg("error querying tag definitions, skipping org tag creation")

					continue
				}

				if exists {
					continue
				}

				input := generated.CreateTagDefinitionInput{
					Name:    tag,
					OwnerID: &orgID,
				}

				if err := mut.Client().TagDefinition.Create().
					SetInput(input).
					Exec(ctx); err != nil {
					logx.FromContext(ctx).Error().Err(err).Str("tag", tag).Msg("error creating tag definition")

					return nil, err
				}
			}

			// continue with the rest of the mutation
			return next.Mutate(ctx, m)
		})
	}, hook.HasOp(ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne))
}
