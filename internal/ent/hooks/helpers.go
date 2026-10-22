package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/riverqueue/river"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/riverboat/pkg/riverqueue"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/privacy/utils"
	"github.com/theopenlane/core/v2/internal/workflows/engine"
	"github.com/theopenlane/core/v2/pkg/middleware/transaction"
	"github.com/theopenlane/logx"
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

// getMutationIDs retrieves the IDs from the mutation, handling both single and multiple ID cases
func getMutationIDs(ctx context.Context, m utils.GenericMutation) ([]string, error) {
	objID, ok := m.ID()
	if ok && objID != "" {
		return []string{objID}, nil
	}

	objIDs, err := m.IDs(ctx)
	return objIDs, err
}

// getSingleMutationID retrieves a single ID from the mutation, returning empty string and false if
// zero or multiple IDs are present. This is a convenience wrapper around getMutationIDs for hooks
// that only operate on single-entity mutations.
func getSingleMutationID(ctx context.Context, m utils.GenericMutation) (string, bool, error) {
	ids, err := getMutationIDs(ctx, m)
	if err != nil {
		return "", false, err
	}
	if len(ids) != 1 || ids[0] == "" {
		return "", false, nil
	}

	return ids[0], true, nil
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

// workflowEngineEnabled reports whether the process-wide workflow engine is registered
func workflowEngineEnabled() bool {
	return engine.Enabled()
}
