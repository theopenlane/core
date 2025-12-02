package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/control"
	"github.com/theopenlane/ent/generated/hook"
	"github.com/theopenlane/ent/generated/subcontrol"
	"github.com/theopenlane/ent/privacy/utils"
)

type controlEdgeMutation interface {
	utils.GenericMutation

	ControlsIDs() []string
}

type subcontrolEdgeMutation interface {
	utils.GenericMutation

	SubcontrolsIDs() []string
}

// HookSystemOwnedControls runs on mutations to check for system owned controls
// since only view access to a control is required for edges on tasks, evidence, this
// ensures that system owned controls are not linked to org owned objects
func HookSystemOwnedControls() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if isDeleteOp(ctx, m) {
				return next.Mutate(ctx, m)
			}

			mut, ok := m.(controlEdgeMutation)
			if ok {
				if err := checkControlEdges(ctx, mut); err != nil {
					return nil, err
				}
			}

			subMut, ok := m.(subcontrolEdgeMutation)
			if ok {
				if err := checkSubcontrolEdges(ctx, subMut); err != nil {
					return nil, err
				}
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate)
}

// checkControlEdges ensures that system owned (sub)controls are not being linked to edges of org owned objects
func checkControlEdges(ctx context.Context, m controlEdgeMutation) error {
	// ensure system owned controls are not being linked
	controlIDs := m.ControlsIDs()

	if len(controlIDs) > 0 {
		systemOwnedControlCount, err := m.Client().Control.Query().Where(
			control.IDIn(controlIDs...),
			control.SystemOwnedEQ(true),
		).Count(ctx)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msgf("failed to query controls for %s mutation", m.Type())

			return err
		}

		if systemOwnedControlCount > 0 {
			logx.FromContext(ctx).Warn().Msgf("permission denied: attempt to link system owned control to %s", m.Type())

			return generated.ErrPermissionDenied
		}
	}

	return nil
}

func checkSubcontrolEdges(ctx context.Context, m subcontrolEdgeMutation) error {
	// ensure system owned controls are not being linked
	subcontrolIDs := m.SubcontrolsIDs()

	if len(subcontrolIDs) > 0 {
		systemOwnedControlCount, err := m.Client().Subcontrol.Query().Where(
			subcontrol.IDIn(subcontrolIDs...),
			subcontrol.SystemOwnedEQ(true),
		).Count(ctx)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msgf("failed to query subcontrols for %s mutation", m.Type())

			return err
		}

		if systemOwnedControlCount > 0 {
			logx.FromContext(ctx).Warn().Msgf("permission denied: attempt to link system owned subcontrol to %s", m.Type())

			return generated.ErrPermissionDenied
		}
	}

	return nil
}
