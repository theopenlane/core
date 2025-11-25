package hooks

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"

	"github.com/gertd/go-pluralize"
	"github.com/stoewer/go-strcase"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/tagdefinition"
	"github.com/theopenlane/core/pkg/logx"
)

var (
	// ErrTagDefinitionInUse is returned when a tag definition is in use and cannot be deleted
	ErrTagDefinitionInUse = errors.New("tag definition is in use")
	// ErrTagDefinitionInUse is returned when there is a db level error fetching all org owned tags
	ErrTagsNotFetched = errors.New("an error occurred while fetching all tags")
)

func HookTagDefintion() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TagDefinitionFunc(func(ctx context.Context, m *generated.TagDefinitionMutation) (generated.Value, error) {
			name, ok := m.Name()
			if !ok {
				return next.Mutate(ctx, m)
			}

			name = strings.TrimSpace(name)

			searchValue := strings.ToLower(name)

			tag, err := m.Client().TagDefinition.Query().
				Where(
					tagdefinition.Or(
						tagdefinition.NameEqualFold(name),
						func(s *sql.Selector) {
							aliasPredicate := sqljson.ValueContains(
								tagdefinition.FieldAliases,
								name,
							)

							if name == searchValue {
								s.Where(aliasPredicate)

								return
							}

							s.Where(
								sql.Or(
									aliasPredicate,
									sqljson.ValueContains(
										tagdefinition.FieldAliases,
										searchValue,
									),
								),
							)
						},
					),
				).
				First(ctx)

			// if not found, create the tag
			if generated.IsNotFound(err) {

				m.SetSlug(strcase.KebabCase(name))
				m.SetName(searchValue)
				if values, hasAliases := m.Aliases(); hasAliases {
					aliases := make([]string, 0, len(values))

					for _, v := range values {
						alias := strings.TrimSpace(v)
						if alias == "" {
							continue
						}

						aliases = append(aliases, strings.ToLower(alias))
					}

					if len(aliases) > 0 {
						m.SetAliases(aliases)
					}
				}

				return next.Mutate(ctx, m)
			}

			if err != nil {
				logx.FromContext(ctx).Warn().Err(err).Msg("error fetching matching tag definitions")
				return nil, ErrTagsNotFetched
			}

			return tag, nil
		})
	}, ent.OpCreate)
}

// HookTagDefinitionDelete checks if the tag definition(s) being deleted is in use by any workflow definition.
// If in use, the deletion cannot proceed
func HookTagDefinitionDelete() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TagDefinitionFunc(func(ctx context.Context, m *generated.TagDefinitionMutation) (generated.Value, error) {
			if !isDeleteOp(ctx, m) {
				return next.Mutate(ctx, m)
			}

			ids := getMutationIDs(ctx, m)
			if len(ids) == 0 {
				return next.Mutate(ctx, m)
			}

			client := m.Client()
			tagDefs, err := client.TagDefinition.Query().
				Where(tagdefinition.IDIn(ids...)).
				Select(
					tagdefinition.FieldID,
					tagdefinition.FieldName,
				).
				All(ctx)
			if err != nil {
				return nil, err
			}

			var errs []string
			var mu sync.Mutex

			funcs := make([]func(), 0)
			for _, tagDef := range tagDefs {
				funcs = append(funcs, isTagDefinitionInUse(ctx, client, tagDef.ID, tagDef.Name, &errs, &mu)...)
			}

			if len(funcs) == 0 {
				return next.Mutate(ctx, m)
			}

			client.PondPool.SubmitMultipleAndWait(funcs)

			if len(errs) > 0 {
				logx.FromContext(ctx).Error().
					Int("error_count", len(errs)).
					Strs("errors", errs).
					Msg("tag definition deletion failed: tag definitions are in use")
				return nil, fmt.Errorf("%w: %d tag definition(s) are in use and cannot be deleted", ErrTagDefinitionInUse, len(errs)) //nolint:err113
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpDeleteOne|ent.OpDelete|ent.OpUpdateOne|ent.OpUpdate)
}

// isTagDefinitionInUse checks if a tag definition is in use by any schema.
//
// It is just workflow for now but can always be extended. See hooks/customenums.go
func isTagDefinitionInUse(ctx context.Context, client *generated.Client, tagDefID, tagName string, allErrors *[]string, mu *sync.Mutex) []func() {
	type tableConfig struct {
		table           string
		field           string
		referencedTable string
		label           string
	}

	configs := []tableConfig{
		{
			table:           "tag_definitions",
			field:           "workflow_definition_tag_definitions",
			referencedTable: "workflow_definitions",
			label:           "workflow definition",
		},
	}

	funcs := make([]func(), 0, len(configs))

	for _, config := range configs {
		funcs = append(funcs, func() {
			// we need to check if tag_definition is in use by checking the referenced entity and making sure it exists and
			// not deleted
			query := fmt.Sprintf(`
				SELECT count(td.id) 
				FROM %s td
				INNER JOIN %s ref ON td.%s = ref.id
				WHERE td.id = $1 
					AND td.deleted_at IS NULL 
					AND ref.deleted_at IS NULL`, config.table, config.referencedTable, config.field)

			var rows sql.Rows
			if err := client.Driver().Query(ctx, query, []any{tagDefID}, &rows); err != nil {
				mu.Lock()
				logx.FromContext(ctx).Error().Err(err).
					Str("table", config.table).
					Str("field", config.field).
					Str("tag_definition_id", tagDefID).
					Msg("failed to query tag definition edges")
				*allErrors = append(*allErrors, fmt.Sprintf("failed to check if tag definition %s is in use: %v", tagName, err))
				mu.Unlock()
				return
			}
			defer rows.Close()

			var count int
			if rows.Next() {
				if err := rows.Scan(&count); err != nil {
					mu.Lock()
					logx.FromContext(ctx).Error().Err(err).
						Str("table", config.table).
						Str("field", config.field).
						Str("tag_definition_id", tagDefID).
						Msg("failed to scan tag definition edge count")
					*allErrors = append(*allErrors, fmt.Sprintf("failed to check if tag definition %s is in use: %v", tagName, err))
					mu.Unlock()
					return
				}
			}

			if count > 0 {
				mu.Lock()

				label := config.label
				if count != 1 {
					label = pluralize.NewClient().Plural(config.label)
				}

				*allErrors = append(*allErrors, fmt.Sprintf("the tag definition %s is in use by %d %s and cannot be deleted until those are updated",
					tagName, count, label))
				mu.Unlock()
			}
		})
	}

	return funcs
}
