package hooks

import (
	"context"
	"fmt"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/mappedcontrol"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/iam/auth"
)

// HookMappedControl runs on mapped control create and update mutations
func HookMappedControl() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return hook.MappedControlFunc(func(ctx context.Context, m *generated.MappedControlMutation) (generated.Value, error) {
			mc, ok := m.Source()
			if ok && mc == enums.MappingSourceSuggested {
				// ensure user is a system admin
				if !auth.IsSystemAdminFromContext(ctx) {
					return nil, fmt.Errorf("%w: only system admins can create suggested mappings", ErrInvalidInput)
				}

			}

			return next.Mutate(ctx, m)
		})
	},
		hook.And(
			hook.HasFields(mappedcontrol.FieldSource),
			hook.HasOp(ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate),
		),
	)
}
