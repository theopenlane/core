package hooks

import (
	"context"
	"fmt"

	"entgo.io/ent"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/template"
	"github.com/theopenlane/core/pkg/objects"
)

type clearingNDAFilesKeyOp struct{}

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
			if ctx.Value(clearingNDAFilesKeyOp{}) != nil {
				return next.Mutate(ctx, m)
			}

			// check for uploaded files
			fileIDs := objects.GetFileIDsFromContext(ctx)
			if len(fileIDs) > 0 {
				var err error

				ctx, err = checkTemplateFiles(ctx, m)
				if err != nil {
					return nil, err
				}

				if trustCenterID, ok := m.TrustCenterID(); ok && trustCenterID != "" {
					clearCtx := context.WithValue(ctx, clearingNDAFilesKeyOp{}, true)
					_, err = m.Client().Template.Update().
						Where(template.TrustCenterIDEQ(trustCenterID)).
						ClearFiles().
						Save(clearCtx)
					if err != nil {
						return nil, err
					}
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
