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
			controls, err := sc.Controls(allowCtx)
			if err != nil {
				return retVal, err
			}

			if len(controls) == 0 {
				return nil, ErrNoControls
			}

			return retVal, nil
		})
	}, ent.OpUpdate|ent.OpUpdateOne)
}
