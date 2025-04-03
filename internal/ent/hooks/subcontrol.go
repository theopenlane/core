package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
)

// HookSubcontrolUpdate ensures that there is at least 1 control assigned to the subcontrol
func HookSubcontrolUpdate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.SubcontrolFunc(func(ctx context.Context, m *generated.SubcontrolMutation) (generated.Value, error) {
			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return retVal, err
			}

			// do not check if the operation is a delete operation (soft delete)
			if isDeleteOp(ctx, m) {
				return retVal, nil
			}

			scID, ok := m.ID()
			if !ok {
				return retVal, nil
			}

			sc, err := m.Client().Subcontrol.Get(ctx, scID)
			if err != nil {
				return retVal, err
			}

			// ensure that the subcontrol has at least one control assigned
			allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

			control, err := sc.Control(allowCtx)
			if err != nil {
				return retVal, err
			}

			if control == nil {
				return nil, ErrNoControls
			}

			return retVal, nil
		})
	}, ent.OpUpdate|ent.OpUpdateOne)
}

// HookSubcontrolCreate sets default values for the subcontrol on creation
func HookSubcontrolCreate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.SubcontrolFunc(func(ctx context.Context, m *generated.SubcontrolMutation) (generated.Value, error) {
			// check if the subcontrol has an owner assigned
			if _, ok := m.OwnerID(); !ok {
				// if not, check if the parent control has an owner
				// subcontrols must have a control assigned
				controlID, _ := m.ControlID()

				control, err := m.Client().Control.Get(ctx, controlID)
				if err != nil {
					return nil, err
				}

				// if the control has an owner, assign it to the subcontrol
				if control.OwnerID != "" {
					m.SetOwnerID(control.OwnerID)
				}
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
}
