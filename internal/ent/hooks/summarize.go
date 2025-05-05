package hooks

import (
	"context"
	"errors"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
)

type detailsMutation interface {
	SetSummary(string)
	Details() (string, bool)
	Client() *generated.Client
}

// HookSummarizeDetails summarizes the policy and produces a short human readable copy
func HookSummarizeDetails() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			mut := m.(detailsMutation)

			details, ok := mut.Details()
			if ok && details != "" {
				return next.Mutate(ctx, m)
			}

			summarizer := mut.Client().Summarizer
			if summarizer == nil {
				return nil, errors.New("summarizer client not found") //nolint:err113
			}

			summary, err := summarizer.Summarize(ctx, details)
			if err != nil {
				return nil, err
			}

			mut.SetSummary(summary)
			return next.Mutate(ctx, m)
		})
	}, hook.HasOp(ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne))
}
