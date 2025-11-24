package hooks

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"entgo.io/ent"

	"github.com/samber/lo"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/tagdefinition"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/iam/auth"
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
			if err != nil || orgID == "" {
				logx.FromContext(ctx).Debug().Msg("no organization ID in context, skipping tag definition creation")

				return next.Mutate(ctx, m)
			}

			// for each tag, create the tag definition if it does not already exist
			for _, tag := range uniqueTags {
				if tag == "" {
					continue
				}

				exists, err := mut.Client().TagDefinition.Query().
					Where(tagdefinition.NameEqualFold(tag)).
					Exist(ctx)
				if !exists {
					if err := rule.CheckCurrentOrgAccess(ctx, m, fgax.CanEdit); err != nil {
						if errors.Is(err, privacy.Deny) {
							return nil, fmt.Errorf("insufficient permissions to create tag definition '%s': only users with write access can create new tags", tag) //nolint:err113
						}
					}

					input := generated.CreateTagDefinitionInput{
						Name:    tag,
						OwnerID: &orgID,
					}

					if err := mut.Client().TagDefinition.Create().
						SetInput(input).
						Exec(ctx); err != nil {
						if !generated.IsConstraintError(err) {
							logx.FromContext(ctx).Warn().Err(err).Str("tag", tag).Msg("error creating tag definition")
						}

						// else, another process created it, so we can ignore the error
						logx.FromContext(ctx).Debug().Str("tag", tag).Msg("tag definition already exists, skipping creation")
					}
				} else if err != nil {
					logx.FromContext(ctx).Warn().Err(err).Msg("error querying tag definitions, skipping org tag creation")
				}
			}

			// continue with the rest of the mutation
			return next.Mutate(ctx, m)
		})
	}, hook.HasOp(ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne))
}
