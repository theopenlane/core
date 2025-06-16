package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/control"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/subcontrol"
)

// HookControlReferenceFramework runs on control mutations to set the reference framework
// based on the standard's short name
func HookControlReferenceFramework() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.ControlFunc(func(ctx context.Context, m *generated.ControlMutation) (generated.Value, error) {
			stdCleared := m.StandardIDCleared()
			if stdCleared {
				m.ClearReferenceFramework()

				return next.Mutate(ctx, m)
			}

			standardID, ok := m.StandardID()
			if ok {
				std, err := m.Client().Standard.Get(ctx, standardID)
				if err != nil {
					return nil, err
				}

				m.SetReferenceFramework(std.ShortName)

				if m.Op().Is(ent.OpUpdateOne) {
					id, ok := m.ID()
					if !ok {
						return next.Mutate(ctx, m)
					}

					// set the reference framework on all subcontrols as well
					err = m.Client().Subcontrol.Update().
						Where(subcontrol.HasControlWith(control.ID(id))).
						SetReferenceFramework(std.ShortName).
						Exec(ctx)
					if err != nil {
						return nil, err
					}
				}
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne)
}
