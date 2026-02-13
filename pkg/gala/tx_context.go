package gala

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/theopenlane/utils/contextx"
)

// contextPGXTx stores a pgx transaction for durable dispatch paths that support InsertTx
type contextPGXTx struct {
	Tx pgx.Tx
}

// WithPGXTx stores a pgx transaction in context for durable dispatchers
func WithPGXTx(ctx context.Context, tx pgx.Tx) context.Context {
	return contextx.With(ctx, contextPGXTx{Tx: tx})
}

// PGXTxFromContext returns a pgx transaction when present in context
func PGXTxFromContext(ctx context.Context) (pgx.Tx, bool) {
	stored, ok := contextx.From[contextPGXTx](ctx)
	if !ok || stored.Tx == nil {
		return nil, false
	}

	return stored.Tx, true
}
