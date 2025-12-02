package hooks

import (
	"context"
	"time"

	"entgo.io/ent"

	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/hook"
)

// HookControlImplementation sets default values for the control implementation
func HookControlImplementation() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.ControlImplementationFunc(func(ctx context.Context, m *generated.ControlImplementationMutation) (generated.Value, error) {
			verified, _ := m.Verified()
			verifiedDated, _ := m.VerificationDate()

			if verified && verifiedDated.IsZero() {
				if m.Op() == ent.OpUpdateOne {
					oldDate, err := m.OldVerificationDate(ctx)
					if err != nil {
						return nil, err
					}

					if oldDate.IsZero() {
						// set the verification date to now
						m.SetVerificationDate(time.Now())
					}
				} else {
					m.SetVerificationDate(time.Now())
				}
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne)
}
