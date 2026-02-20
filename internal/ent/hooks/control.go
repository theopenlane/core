package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/control"
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
			if m.Op().Is(ent.OpUpdateOne) {
				return handleControlVisibilityUpdate(ctx, m, next)
			}

			if m.Op().Is(ent.OpUpdate) {
				return handleControlVisibilityUpdate(ctx, m, next)
			}

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
				_, err := handleControlVisibilityCreate(ctx, m, ctrl)
				return v, err
			}

			return v, nil
		})
	}, ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne)
}

// ControlVisibilityTupleAction determines whether wildcard viewer tuples should be
// written or deleted based on the trust center visibility state of a control.
// The caller is responsible for ensuring this is only called for trust center controls.
func ControlVisibilityTupleAction(newVisibility, oldVisibility enums.TrustCenterControlVisibility, visibilityChanged bool) (shouldWrite, shouldDelete bool) {
	if !visibilityChanged {
		return false, false
	}

	if newVisibility == oldVisibility {
		return false, false
	}

	if newVisibility == enums.TrustCenterControlVisibilityPubliclyVisible {
		return true, false
	}

	if oldVisibility == enums.TrustCenterControlVisibilityPubliclyVisible {
		return false, true
	}

	return false, false
}

// handleControlVisibilityCreate adds wildcard viewer tuples on create if the control is publicly visible
func handleControlVisibilityCreate(ctx context.Context, m *generated.ControlMutation, ctrl *generated.Control) (generated.Value, error) {
	shouldWrite, _ := ControlVisibilityTupleAction(
		ctrl.TrustCenterVisibility,
		enums.TrustCenterControlVisibilityNotVisible,
		ctrl.TrustCenterVisibility != enums.TrustCenterControlVisibilityNotVisible,
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

// controlVisibilityTupleChanges returns tuple writes/deletes for a control visibility transition
func controlVisibilityTupleChanges(controlID string, oldVisibility, newVisibility enums.TrustCenterControlVisibility) (writes, deletes []fgax.TupleKey) {
	shouldWrite, shouldDelete := ControlVisibilityTupleAction(newVisibility, oldVisibility, true)

	if shouldWrite {
		writes = fgax.CreateWildcardViewerTuple(controlID, generated.TypeControl)
	}

	if shouldDelete {
		deletes = fgax.CreateWildcardViewerTuple(controlID, generated.TypeControl)
	}

	return writes, deletes
}

// controlVisibilityTupleChangesFromTupleState computes tuple writes/deletes when only the
// current wildcard tuple state is known
func controlVisibilityTupleChangesFromTupleState(controlID string, newVisibility enums.TrustCenterControlVisibility, tupleExists bool) (writes, deletes []fgax.TupleKey) {
	if newVisibility == enums.TrustCenterControlVisibilityPubliclyVisible {
		if !tupleExists {
			writes = fgax.CreateWildcardViewerTuple(controlID, generated.TypeControl)
		}

		return writes, nil
	}

	if tupleExists {
		deletes = fgax.CreateWildcardViewerTuple(controlID, generated.TypeControl)
	}

	return nil, deletes
}

// hasControlWildcardViewerTuple checks whether the wildcard viewer tuple currently exists for a control
func hasControlWildcardViewerTuple(ctx context.Context, m *generated.ControlMutation, controlID string) (bool, error) {
	return m.Authz.CheckAccessHighConsistency(ctx, fgax.AccessCheck{
		SubjectID:   fgax.Wildcard,
		SubjectType: auth.UserSubjectType,
		Relation:    fgax.CanView,
		ObjectID:    controlID,
		ObjectType:  fgax.Kind(generated.TypeControl),
	})
}

type controlVisibilityTransition struct {
	newVisibility    enums.TrustCenterControlVisibility
	oldVisibility    enums.TrustCenterControlVisibility
	hasOldVisibility bool
}

// handleControlVisibilityUpdate manages wildcard viewer tuples for both single and bulk updates.
func handleControlVisibilityUpdate(ctx context.Context, m *generated.ControlMutation, next ent.Mutator) (generated.Value, error) {
	if m.Op().Is(ent.OpUpdateOne) {
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

		transition := controlVisibilityTransition{
			newVisibility: ctrl.TrustCenterVisibility,
		}

		if oldVisibility, oldErr := m.OldTrustCenterVisibility(ctx); oldErr == nil {
			transition.oldVisibility = oldVisibility
			transition.hasOldVisibility = true
		}

		if err := applyControlVisibilityTransitions(ctx, m, map[string]controlVisibilityTransition{
			ctrl.ID: transition,
		}); err != nil {
			return nil, err
		}

		return v, nil
	}

	if m.Op().Is(ent.OpUpdate) {
		visibility, visibilityChanged := m.TrustCenterVisibility()
		if !visibilityChanged {
			return next.Mutate(ctx, m)
		}

		updatedIDs := getMutationIDs(ctx, m)
		if len(updatedIDs) == 0 {
			return next.Mutate(ctx, m)
		}

		oldControls, err := m.Client().Control.Query().
			Where(control.IDIn(updatedIDs...)).
			Select(control.FieldID, control.FieldTrustCenterVisibility, control.FieldIsTrustCenterControl).
			All(ctx)
		if err != nil {
			return nil, err
		}

		oldVisibilityByID := make(map[string]enums.TrustCenterControlVisibility, len(oldControls))
		for _, ctrl := range oldControls {
			if !ctrl.IsTrustCenterControl {
				continue
			}

			oldVisibilityByID[ctrl.ID] = ctrl.TrustCenterVisibility
		}

		v, err := next.Mutate(ctx, m)
		if err != nil {
			return v, err
		}

		if len(oldVisibilityByID) == 0 {
			return v, nil
		}

		transitions := make(map[string]controlVisibilityTransition, len(oldVisibilityByID))
		for controlID, oldVisibility := range oldVisibilityByID {
			transitions[controlID] = controlVisibilityTransition{
				newVisibility:    visibility,
				oldVisibility:    oldVisibility,
				hasOldVisibility: true,
			}
		}

		if err := applyControlVisibilityTransitions(ctx, m, transitions); err != nil {
			return nil, err
		}

		return v, nil
	}

	return next.Mutate(ctx, m)
}

// applyControlVisibilityTransitions applies the necessary wildcard viewer tuple writes/deletes for a set of control visibility transitions
func applyControlVisibilityTransitions(ctx context.Context, m *generated.ControlMutation, transitions map[string]controlVisibilityTransition) error {
	if len(transitions) == 0 {
		return nil
	}

	var writes, deletes []fgax.TupleKey

	for controlID, transition := range transitions {
		var ctrlWrites, ctrlDeletes []fgax.TupleKey

		if transition.hasOldVisibility {
			ctrlWrites, ctrlDeletes = controlVisibilityTupleChanges(controlID, transition.oldVisibility, transition.newVisibility)
		} else {
			tupleExists, err := hasControlWildcardViewerTuple(ctx, m, controlID)
			if err != nil {
				return err
			}

			ctrlWrites, ctrlDeletes = controlVisibilityTupleChangesFromTupleState(controlID, transition.newVisibility, tupleExists)
		}

		writes = append(writes, ctrlWrites...)
		deletes = append(deletes, ctrlDeletes...)
	}

	if len(writes) == 0 && len(deletes) == 0 {
		return nil
	}

	_, err := m.Authz.WriteTupleKeys(ctx, writes, deletes)

	return err
}
