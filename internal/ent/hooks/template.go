package hooks

import (
	"context"
	"fmt"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/template"
	"github.com/theopenlane/core/pkg/enums"
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
