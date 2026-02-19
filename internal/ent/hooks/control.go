package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/standard"
	"github.com/theopenlane/core/internal/ent/generated/subcontrol"
	"github.com/theopenlane/core/pkg/logx"
)

// HookControlReferenceFramework runs on control mutations to set the reference framework
// based on the standard's short name
func HookControlReferenceFramework() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.ControlFunc(func(ctx context.Context, m *generated.ControlMutation) (generated.Value, error) {
			shortName := ""
			revision := ""

			// if the control is created with a reference framework, we will use it
			if m.Op().Is(ent.OpCreate) {
				refFramework, ok := m.ReferenceFramework()
				if ok && refFramework != "" {
					// if the reference framework is set, we will use it
					return next.Mutate(ctx, m)
				}
			}

			stdCleared := m.StandardIDCleared()
			if stdCleared {
				m.ClearReferenceFramework()
			} else {
				standardID, ok := m.StandardID()
				if ok {
					std, err := m.Client().Standard.Query().Select(standard.FieldShortName, standard.FieldRevision).
						Where(standard.ID(standardID)).Only(ctx)
					if err != nil {
						return nil, err
					}

					m.SetReferenceFramework(std.ShortName)
					m.SetReferenceFrameworkRevision(std.Revision)
					shortName = std.ShortName
					revision = std.Revision
				}
			}

			// if this is an update and the standard was cleared or the standard was set,
			// we need to update the subcontrols as well
			// this is because the subcontrols inherit the reference framework from the control
			if m.Op().Is(ent.OpUpdateOne) && (stdCleared || shortName != "") {
				id, ok := m.ID()
				if !ok {
					return next.Mutate(ctx, m)
				}

				// allow the subcontrol mutation to run
				// if a user can edit the control, they can edit the subcontrols
				allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

				mut := m.Client().Subcontrol.Update().
					Where(subcontrol.ControlID(id))

				if stdCleared {
					mut.ClearReferenceFramework()
					mut.ClearReferenceFrameworkRevision()
				} else {
					mut.SetReferenceFramework(shortName)
					mut.SetReferenceFrameworkRevision(revision)
				}
				// set the reference framework on all subcontrols as well
				if err := mut.Exec(allowCtx); err != nil {
					return nil, err
				}
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne)
}

// HookControlTrustCenterVisibility manages FGA wildcard viewer tuples when the
// trust_center_visibility field changes on a control, enabling or revoking
// anonymous public access based on the visibility state
func HookControlTrustCenterVisibility() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.ControlFunc(func(ctx context.Context, m *generated.ControlMutation) (generated.Value, error) {
			v, err := next.Mutate(ctx, m)
			if err != nil {
				return v, err
			}

			ctrl, ok := v.(*generated.Control)
			if !ok {
				return v, nil
			}

			// only manage wildcard tuples for trust center controls
			if !ctrl.IsTrustCenterControl {
				return v, nil
			}

			if m.Op().Is(ent.OpCreate) {
				return handleControlVisibilityCreate(ctx, m, ctrl)
			}

			return handleControlVisibilityUpdate(ctx, m, ctrl)
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}

// ControlVisibilityTupleAction determines whether wildcard viewer tuples should be
// written or deleted based on the trust center visibility state of a control
func ControlVisibilityTupleAction(isTrustCenterControl bool, newVisibility, oldVisibility enums.TrustCenterDocumentVisibility, visibilityChanged bool) (shouldWrite, shouldDelete bool) {
	if !isTrustCenterControl {
		return false, false
	}

	if !visibilityChanged {
		return false, false
	}

	if newVisibility == oldVisibility {
		return false, false
	}

	if newVisibility == enums.TrustCenterDocumentVisibilityPubliclyVisible {
		return true, false
	}

	if oldVisibility == enums.TrustCenterDocumentVisibilityPubliclyVisible {
		return false, true
	}

	return false, false
}

// handleControlVisibilityCreate adds wildcard viewer tuples on create if the control is publicly visible
func handleControlVisibilityCreate(ctx context.Context, m *generated.ControlMutation, ctrl *generated.Control) (generated.Value, error) {
	shouldWrite, _ := ControlVisibilityTupleAction(
		ctrl.IsTrustCenterControl,
		ctrl.TrustCenterVisibility,
		enums.TrustCenterDocumentVisibilityNotVisible,
		ctrl.TrustCenterVisibility != enums.TrustCenterDocumentVisibilityNotVisible,
	)

	if !shouldWrite {
		return ctrl, nil
	}

	tuples := fgax.CreateWildcardViewerTuple(ctrl.ID, generated.TypeControl)

	logx.FromContext(ctx).Debug().Str("control_id", ctrl.ID).Msg("creating wildcard viewer tuples for publicly visible trust center control")

	if _, err := m.Authz.WriteTupleKeys(ctx, tuples, nil); err != nil {
		return nil, err
	}

	return ctrl, nil
}

// handleControlVisibilityUpdate manages wildcard viewer tuples when visibility changes on update
func handleControlVisibilityUpdate(ctx context.Context, m *generated.ControlMutation, ctrl *generated.Control) (generated.Value, error) {
	visibility, visibilityChanged := m.TrustCenterVisibility()
	if !visibilityChanged {
		return ctrl, nil
	}

	oldVisibility, err := m.OldTrustCenterVisibility(ctx)
	if err != nil {
		return nil, err
	}

	shouldWrite, shouldDelete := ControlVisibilityTupleAction(ctrl.IsTrustCenterControl, visibility, oldVisibility, true)

	var writes, deletes []fgax.TupleKey

	if shouldWrite {
		writes = fgax.CreateWildcardViewerTuple(ctrl.ID, generated.TypeControl)
	}

	if shouldDelete {
		deletes = fgax.CreateWildcardViewerTuple(ctrl.ID, generated.TypeControl)
	}

	if len(writes) > 0 || len(deletes) > 0 {
		logx.FromContext(ctx).Debug().Str("control_id", ctrl.ID).Str("old_visibility", string(oldVisibility)).Str("new_visibility", string(visibility)).Msg("updating wildcard viewer tuples for trust center control visibility change")

		if _, err := m.Authz.WriteTupleKeys(ctx, writes, deletes); err != nil {
			return nil, err
		}
	}

	return ctrl, nil
}
