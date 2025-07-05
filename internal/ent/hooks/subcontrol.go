package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/rs/zerolog/log"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/control"
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
			// check if the subcontrol has an owner organization assigned
			c, err := getParentControl(ctx, m)
			if err != nil {
				return nil, err
			}

			if c != nil {
				if c.ReferenceFramework != nil {
					m.SetReferenceFramework(*c.ReferenceFramework)
				}

				if c.ControlOwnerID != nil && *c.ControlOwnerID != "" {
					m.SetControlOwnerID(*c.ControlOwnerID)
				}
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne)
}

// getParentControl retrieves the parent control of a subcontrol mutation
// if returns the control reference framework and control owner if they are set
// if the control ID is not set, it returns nil
func getParentControl(ctx context.Context, m *generated.SubcontrolMutation) (*generated.Control, error) {
	controlID, ok := m.ControlID()
	if !ok || controlID == "" {
		return nil, nil
	}

	fields := []string{}

	referenceFramework, ok := m.ReferenceFramework()
	if !ok || referenceFramework == "" {
		// if the reference framework is not set, we will get it from the parent control
		fields = append(fields, control.FieldReferenceFramework)
	}

	controlOwnerID, ok := m.ControlOwnerID()
	if !ok {
		// if the control owner is not set, we will get it from the parent control
		// this is the group that owns the control, it is used for task assignment,
		// approval, etc.
		fields = append(fields, control.FieldControlOwnerID)
		log.Warn().Msg("checking parent control for value")
	}

	// if the controlOwner was explicitly set to an empty string, clear it
	// and don't check for the parent control owner
	if controlOwnerID == "" {
		log.Warn().Msg("clearing control owner field")
		m.ClearControlOwnerID()
	}

	// doing to do now
	if len(fields) == 0 {
		return nil, nil
	}

	c, err := m.Client().Control.Query().
		Where(control.ID(controlID)).
		Select(fields...).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	return c, nil
}
