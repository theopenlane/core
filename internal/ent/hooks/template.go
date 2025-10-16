package hooks

import (
	"context"
	"fmt"

	"entgo.io/ent"

	"github.com/rs/zerolog"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/template"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
)

// HookTemplate runs on template create and update mutations
func HookTemplate() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return hook.TemplateFunc(func(ctx context.Context, m *generated.TemplateMutation) (generated.Value, error) {
			tt, ok := m.TemplateType()
			if ok && tt == enums.RootTemplate {
				// ensure user is a system admin
				if !auth.IsSystemAdminFromContext(ctx) {
					return nil, fmt.Errorf("%w: only system admins can create or update root templates", ErrInvalidInput)
				}

			}

			return next.Mutate(ctx, m)
		})
	},
		hook.And(
			hook.HasFields(template.FieldTemplateType),
			hook.HasOp(ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate),
		),
	)
}

func HookTemplateAuthz() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return hook.TemplateFunc(func(ctx context.Context, m *generated.TemplateMutation) (ent.Value, error) {
			// do the mutation, and then create/delete the relationship
			retValue, err := next.Mutate(ctx, m)
			if err != nil {
				// if we error, do not attempt to create the relationships
				return retValue, err
			}

			if m.Op().Is(ent.OpCreate) {
				createdTemplate := retValue.(*generated.Template)
				if createdTemplate.SystemOwned {
					err = templateCreateHook(ctx, m)
				}
			} else if isDeleteOp(ctx, m) {
				err = templateDeleteHook(ctx, m)
			}

			return retValue, err
		})
	}
}

func HookTemplateFiles() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TemplateFunc(func(ctx context.Context, m *generated.TemplateMutation) (generated.Value, error) {
			// check for uploaded files
			fileIDs := objects.GetFileIDsFromContext(ctx)
			if len(fileIDs) > 0 {
				var err error

				ctx, err = checkTemplateFiles(ctx, m)
				if err != nil {
					return nil, err
				}

				m.AddFileIDs(fileIDs...)
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate)
}

func checkTemplateFiles(ctx context.Context, m *generated.TemplateMutation) (context.Context, error) {
	key := "templateFiles"

	files, _ := objects.FilesFromContextWithKey(ctx, key)
	if len(files) == 0 {
		return ctx, nil
	}

	adapter := objects.NewGenericMutationAdapter(m,
		func(mut *generated.TemplateMutation) (string, bool) { return mut.ID() },
		func(mut *generated.TemplateMutation) string { return mut.Type() },
	)

	return objects.ProcessFilesForMutation(ctx, adapter, key)
}

func templateCreateHook(ctx context.Context, m *generated.TemplateMutation) error {
	objID, exists := m.ID()
	if !exists {
		return nil
	}

	wildcardTuple := fgax.CreateWildcardViewerTuple(objID, generated.TypeTemplate)
	zerolog.Ctx(ctx).Debug().Interface("request", wildcardTuple).
		Msg("creating public viewer relationship tuples")

	if _, err := m.Authz.WriteTupleKeys(ctx, wildcardTuple, nil); err != nil {
		return err
	}

	return nil
}

func templateDeleteHook(ctx context.Context, m *generated.TemplateMutation) error {
	objID, ok := m.ID()
	if !ok {
		return nil
	}

	objType := GetObjectTypeFromEntMutation(m)
	object := fmt.Sprintf("%s:%s", objType, objID)

	zerolog.Ctx(ctx).Debug().Str("object", object).Msg("deleting relationship tuples")

	if err := m.Authz.DeleteAllObjectRelations(ctx, object, userRoles); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("failed to delete relationship tuples")

		return ErrInternalServerError
	}

	zerolog.Ctx(ctx).Debug().Str("object", object).Msg("deleted relationship tuples")

	return nil
}
