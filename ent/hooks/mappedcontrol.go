package hooks

import (
	"context"
	"fmt"

	"entgo.io/ent"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/hook"
	"github.com/theopenlane/ent/generated/mappedcontrol"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/shared/enums"
)

// HookMappedControl runs on mapped control create and update mutations to restrict certain fields to system admins only
func HookMappedControl() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return hook.MappedControlFunc(func(ctx context.Context, m *generated.MappedControlMutation) (generated.Value, error) {
			if auth.IsSystemAdminFromContext(ctx) {
				return next.Mutate(ctx, m)
			}

			// only system admins can create suggested mappings
			mc, ok := m.Source()
			if ok && mc == enums.MappingSourceSuggested {
				return nil, fmt.Errorf("%w: only system admins can create suggested mappings", ErrInvalidInput)
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
