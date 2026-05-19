package hooks

import (
	"context"
	"fmt"

	"entgo.io/ent"
	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/assessment"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/template"
	"github.com/theopenlane/core/pkg/objects"
)

var clearingNDAFilesOperationContextKey = contextx.NewKey[bool]()

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

			kind, ok, err := getTemplateKind(ctx, m)
			if err != nil {
				return nil, err
			}

			if !ok || kind != enums.TemplateKindExternalIntake {
				return next.Mutate(ctx, m)
			}

			if !auth.IsSystemAdminFromContext(ctx) {
				return nil, fmt.Errorf("%w: only system admins can create or update a vendor intake template", ErrInvalidInput)
			}

			value, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			tmpl, ok := value.(*generated.Template)
			if !ok {
				return value, nil
			}

			if m.Op().Is(ent.OpCreate) {
				if _, err = m.Client().Assessment.Create().
					SetName(tmpl.Name).
					SetTemplateID(tmpl.ID).
					SetAssessmentType(enums.AssessmentTypeExternal).
					SetSystemOwned(true).
					Save(ctx); err != nil {
					return nil, err
				}

				return value, nil
			}

			if _, err = m.Client().Assessment.Update().
				Where(
					assessment.TemplateIDEQ(tmpl.ID),
					assessment.SystemOwnedEQ(true),
					assessment.AssessmentTypeEQ(enums.AssessmentTypeExternal),
				).
				SetName(tmpl.Name).
				SetTemplateID(tmpl.ID).
				Save(ctx); err != nil {
				return nil, err
			}

			return value, nil

		})
	},
		hook.And(
			hook.Or(
				hook.HasFields(template.FieldTemplateType),
				hook.HasFields(template.FieldKind),
				hook.HasFields(template.FieldName),
				hook.HasFields(template.FieldJsonconfig),
				hook.HasFields(template.FieldUischema),
			),
			hook.HasOp(ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate),
		),
	)
}

func getTemplateKind(ctx context.Context, m *generated.TemplateMutation) (enums.TemplateKind, bool, error) {
	if kind, ok := m.Kind(); ok {
		return kind, true, nil
	}

	if !m.Op().Is(ent.OpUpdateOne) {
		return "", false, nil
	}

	kind, err := m.OldKind(ctx)
	if err != nil {
		return "", false, err
	}

	return kind, true, nil
}

func HookTemplateFiles() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TemplateFunc(func(ctx context.Context, m *generated.TemplateMutation) (generated.Value, error) {
			if skipping, ok := clearingNDAFilesOperationContextKey.Get(ctx); ok && skipping {
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
					clearCtx := clearingNDAFilesOperationContextKey.Set(ctx, true)
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

	return objects.ProcessFilesForMutation(ctx, m, key)
}
