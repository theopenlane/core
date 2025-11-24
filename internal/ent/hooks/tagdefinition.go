package hooks

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"entgo.io/ent"

	"github.com/stoewer/go-strcase"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/tagdefinition"
	"github.com/theopenlane/core/internal/ent/generated/workflowdefinition"
	"github.com/theopenlane/core/pkg/logx"
)

func sluggify(s string) string {
	return strings.ReplaceAll(
		strcase.SnakeCase(
			strings.ToLower(s)),
		"_", "-")
}

var (
	// ErrTagDefinitionInUse is returned when a tag definition is in use and cannot be deleted
	ErrTagDefinitionInUse = errors.New("tag definition is in use")
)

func HookTagDefintion() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TagDefinitionFunc(func(ctx context.Context, m *generated.TagDefinitionMutation) (generated.Value, error) {
			name, ok := m.Name()
			if !ok {
				return next.Mutate(ctx, m)
			}

			tagDefs, err := m.Client().TagDefinition.Query().
				All(ctx)
			if err != nil {
				logx.FromContext(ctx).Warn().Err(err).Msg("error fetching all org tags")
				return nil, errors.New("an error occurred while fetching all tags") //nolint:err113
			}

			for _, tagDef := range tagDefs {
				if strings.EqualFold(tagDef.Name, name) {
					// tag with this exact name already exists, return it instead of creating a duplicate
					return tagDef, nil
				}

				for _, alias := range tagDef.Aliases {
					if strings.EqualFold(alias, name) {
						// we found a match - the provided name is an alias of an existing tag
						// so we should not create a new tag definition
						// return the existing tag definition instead of creating a new one
						return tagDef, nil
					}
				}
			}

			m.SetSlug(sluggify(name))

			return next.Mutate(ctx, m)
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
	type countTask struct {
		queryFunc func(context.Context) (int, error)
		label     string
	}

	var operations []countTask

	operations = append(operations, countTask{
		queryFunc: func(ctx context.Context) (int, error) {
			return client.WorkflowDefinition.Query().
				Where(workflowdefinition.HasTagDefinitionsWith(tagdefinition.ID(tagDefID))).
				Count(ctx)
		},
		label: "workflow_definition",
	})

	funcs := make([]func(), 0, len(operations))

	for _, op := range operations {
		funcs = append(funcs, func() {
			if count, err := op.queryFunc(ctx); err == nil && count > 0 {
				mu.Lock()

				label := op.label
				if count != 1 {
					label = "workflow_definitions"
				}

				*allErrors = append(*allErrors, fmt.Sprintf("the tag definition %s is in use by %d %s and cannot be deleted until those are updated",
					tagName, count, label))
				mu.Unlock()
			}
		})
	}

	return funcs
}
