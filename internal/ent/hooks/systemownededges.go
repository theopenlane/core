package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/control"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/subcontrol"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/logx"
)

type controlEdgeMutation interface {
	utils.GenericMutation

	ControlsIDs() []string
	SubcontrolsIDs() []string
}

// HookSystemOwnedControls runs on evidence mutations to check for system owned controls
// since only view access to a control is required for edges on tasks, evidence, this
// ensures that system owned controls are not linked to org owned objects
func HookSystemOwnedControls() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if isDeleteOp(ctx, m) {
				return next.Mutate(ctx, m)
			}

			mut, ok := m.(controlEdgeMutation)
			if !ok {

				// nothing to do if its not the right mutation type
				return next.Mutate(ctx, m)
			}

			if err := checkControlEdges(ctx, mut); err != nil {
				return nil, err
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
		controlEdges, err := m.Client().Control.Query().Where(
			control.IDIn(controlIDs...),
			control.SystemOwnedEQ(true),
		).All(ctx)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to query controls for evidence mutation")

			return err
		}

		if len(controlEdges) > 0 {
			logx.FromContext(ctx).Warn().Msg("permission denied: attempt to link system owned control to evidence")

			return generated.ErrPermissionDenied
		}
	}

	// ensure system owned controls are not being linked
	subcontrolIDs := m.SubcontrolsIDs()

	if len(subcontrolIDs) > 0 {
		subcontrolEdges, err := m.Client().Subcontrol.Query().Where(
			subcontrol.IDIn(subcontrolIDs...),
			subcontrol.SystemOwnedEQ(true),
		).All(ctx)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to query subcontrols for evidence mutation")
			return err
		}

		if len(subcontrolEdges) > 0 {
			logx.FromContext(ctx).Warn().Msg("permission denied: attempt to link system owned subcontrol to evidence")

			return generated.ErrPermissionDenied
		}
	}

	return nil
}
