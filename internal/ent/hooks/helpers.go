package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/entx"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/middleware/transaction"
)

// GenericMutation is an interface for getting a mutation ID and type
type GenericMutation interface {
	ID() (id string, exists bool)
	Type() string
	Client() *generated.Client
}

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
					log.Debug().Ctx(ctx).Err(err2).Msg("post mutation hook error")
				}
			}

			return v, err
		})
	})
}
