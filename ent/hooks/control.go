package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/hook"
	"github.com/theopenlane/ent/generated/privacy"
	"github.com/theopenlane/ent/generated/standard"
	"github.com/theopenlane/ent/generated/subcontrol"
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
