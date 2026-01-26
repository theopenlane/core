package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/riverqueue/river"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/riverboat/pkg/riverqueue"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/middleware/transaction"
)

// isDeleteOp checks if the mutation is a deletion operation.
// which includes soft delete, delete, and delete one.
func isDeleteOp(ctx context.Context, m ent.Mutation) bool {
	return entx.CheckIsSoftDeleteType(ctx, m.Type()) || m.Op().Is(ent.OpDelete) || m.Op().Is(ent.OpDeleteOne)
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
					logx.FromContext(ctx).Debug().Ctx(ctx).Err(err2).Msg("post mutation hook error")
				}
			}

			return v, err
		})
	})
}

// getMutationIDs retrieves the IDs from the mutation, handling both single and multiple ID cases
func getMutationIDs(ctx context.Context, m utils.GenericMutation) []string {
	objID, ok := m.ID()
	if ok && objID != "" {
		return []string{objID}
	}

	objIDs, err := m.IDs(ctx)
	if err == nil {
		return objIDs
	}

	return nil
}

// enqueueJob inserts a job when a job client is available, otherwise logs and skips.
func enqueueJob(ctx context.Context, jobClient riverqueue.JobClient, args river.JobArgs, opts *river.InsertOpts) error {
	if jobClient == nil {
		logx.FromContext(ctx).Warn().Str("job_kind", "unknown").Msg("job client is nil, skipping job insert")
		return nil
	}

	_, err := jobClient.Insert(ctx, args, opts)

	return err
}

// workflowEngineEnabled reports whether workflows are enabled for this client
func workflowEngineEnabled(client *generated.Client) bool {
	return client != nil && client.WorkflowEngine != nil
}
