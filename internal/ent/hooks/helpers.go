package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/entx"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/middleware/transaction"
)

// isDeleteOp checks if the mutation is a deletion operation.
// which includes soft delete, delete, and delete one.
func isDeleteOp(ctx context.Context, m ent.Mutation) bool {
	return entx.CheckIsSoftDelete(ctx) || m.Op().Is(ent.OpDelete) || m.Op().Is(ent.OpDeleteOne)
}

// transactionFromContext returns the transaction from the context if it exists
func transactionFromContext(ctx context.Context) *generated.Tx {
	// check if the transaction is in the context
	// this is returned from all graphql requests
	tx := generated.TxFromContext(ctx)
	if tx != nil {
		return tx
	}

	// check if the transaction is in the context
	// from the REST middleware
	return transaction.FromContext(ctx)
}

// runtimeHooks is a list of post-mutation hooks that are executed after a mutation operation is performed
var runtimeHooks []func(ent.Mutator) ent.Mutator

// The `AddPostMutationHook` function is used to add a post-mutation hook to the list of runtime hooks.
// This function takes a hook function as a parameter, which will be executed after a mutation
// operation is performed. The hook function is expected to take a context and a value of type `T` as
// input parameters and return an error if any
func AddPostMutationHook[T any](hook func(ctx context.Context, v T) error) {
	runtimeHooks = append(runtimeHooks, func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			v, err := next.Mutate(ctx, m)
			if err != nil {
				return v, err
			}

			entvalue, ok := v.(T)

			if ok {
				err2 := hook(ctx, entvalue)
				if err2 != nil {
					zerolog.Ctx(ctx).Debug().Ctx(ctx).Err(err2).Msg("post mutation hook error")
				}
			}

			return v, err
		})
	})
}

// getMutationIDs retrieves the IDs from the mutation, handling both single and multiple ID cases
func getMutationIDs(ctx context.Context, m utils.GenericMutation) []string {
	switch m.Op() {
	case ent.OpDelete, ent.OpUpdate:
		objIDs, err := m.IDs(ctx)
		if err == nil {
			return objIDs
		}

		log.Error().Err(err).Msg("failed to get IDs from mutation")
	case ent.OpDeleteOne, ent.OpUpdateOne:
		objID, ok := m.ID()
		if ok && objID != "" {
			return []string{objID}
		}

		log.Error().Msg("failed to get object IDs from mutation")
	}

	return nil
}
