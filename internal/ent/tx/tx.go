package tx

import (
	"context"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/pkg/middleware/transaction"
)

// AfterCommit runs fn after the transaction in ctx commits, or immediately when there is none
func AfterCommit(ctx context.Context, fn func()) {
	current := FromContext(ctx)
	if current == nil {
		fn()

		return
	}

	current.OnCommit(func(next generated.Committer) generated.Committer {
		return generated.CommitFunc(func(commitCtx context.Context, committed *generated.Tx) error {
			err := next.Commit(commitCtx, committed)
			if err == nil {
				fn()
			}

			return err
		})
	})
}

// FromContext returns the graphql or REST transaction in ctx, or nil
func FromContext(ctx context.Context) *generated.Tx {
	if current := generated.TxFromContext(ctx); current != nil {
		return current
	}

	return transaction.FromContext(ctx)
}
