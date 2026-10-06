package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/hook"
	"github.com/theopenlane/core/v2/internal/ent/generated/internalpolicy"
)

// HookAssessmentPolicyRevision defaults the policy revision to the internal policy's current revision when not provided
func HookAssessmentPolicyRevision() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.AssessmentPolicyFunc(func(ctx context.Context, m *generated.AssessmentPolicyMutation) (generated.Value, error) {
			if _, ok := m.PolicyRevision(); ok {
				return next.Mutate(ctx, m)
			}

			policyID, ok := m.InternalPolicyID()
			if !ok {
				return next.Mutate(ctx, m)
			}

			policy, err := m.Client().InternalPolicy.Query().
				Where(internalpolicy.ID(policyID)).
				Select(internalpolicy.FieldRevision).
				Only(ctx)
			if err != nil {
				return nil, err
			}

			m.SetPolicyRevision(policy.Revision)

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
}
