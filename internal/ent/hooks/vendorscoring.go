package hooks

import (
	"context"

	"entgo.io/ent"
)

// HookVendorRiskScoreCompute sets the score field based on impact x likelihood.
// Full implementation is populated after code generation.
func HookVendorRiskScoreCompute() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			return next.Mutate(ctx, m)
		})
	}
}

// HookVendorRiskScoreAggregate recomputes Entity.risk_score and Entity.risk_rating
// after a VendorRiskScore is created, updated, or deleted.
// Full implementation is populated after code generation.
func HookVendorRiskScoreAggregate() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			return next.Mutate(ctx, m)
		})
	}
}
