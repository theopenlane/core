package workflows

import (
	"context"
	"errors"
	"fmt"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/workflows/observability"
	"github.com/theopenlane/core/pkg/logx"
)

var (
	// ErrTxBegin indicates the transaction failed to start.
	ErrTxBegin = errors.New("failed to begin transaction")
	// ErrTxCommit indicates the transaction failed to commit.
	ErrTxCommit = errors.New("failed to commit transaction")
)

// TxError wraps a transaction error with a stage marker.
type TxError struct {
	Stage error
	Err   error
}

// Error implements error.
func (e *TxError) Error() string {
	return fmt.Sprintf("%s: %v", e.Stage, e.Err)
}

// Unwrap returns the underlying error.
func (e *TxError) Unwrap() error {
	return e.Err
}

// Is reports whether the error matches the stage sentinel.
func (e *TxError) Is(target error) bool {
	return target == e.Stage
}

// WithTx runs fn inside a transaction and ensures rollback on error.
func WithTx[T any](ctx context.Context, client *generated.Client, scope *observability.Scope, fn func(tx *generated.Tx) (T, error)) (T, error) {
	var zero T

	if client == nil {
		return zero, &TxError{Stage: ErrTxBegin, Err: errors.New("nil client")}
	}

	tx, err := client.Tx(ctx)
	if err != nil {
		return zero, &TxError{Stage: ErrTxBegin, Err: err}
	}

	committed := false
	defer func() {
		if committed {
			return
		}
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			if scope != nil {
				scope.Warn(rollbackErr, nil)
			} else {
				logx.FromContext(ctx).Warn().Err(rollbackErr).Msg("failed to rollback transaction")
			}
		}
	}()

	result, err := fn(tx)
	if err != nil {
		return zero, err
	}

	if err := tx.Commit(); err != nil {
		return zero, &TxError{Stage: ErrTxCommit, Err: err}
	}
	committed = true

	return result, nil
}
