package hooks

import (
	"context"
	"fmt"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/template"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/iam/auth"
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

	// get the file from the context, if it exists
	file, _ := objects.FilesFromContextWithKey(ctx, key)

	// return early if no file is provided
	if file == nil {
		return ctx, nil
	}

	// we should only have one file
	if len(file) > 1 {
		return ctx, ErrNotSingularUpload
	}

	// this should always be true, but check just in case
	if file[0].FieldName == key {
		file[0].Parent.ID, _ = m.ID()
		file[0].Parent.Type = m.Type()

		ctx = objects.UpdateFileInContextByKey(ctx, key, file[0])
	}

	return ctx, nil
}
