//go:build test

package hooks_test

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	enttx "github.com/theopenlane/core/v2/internal/ent/tx"
	"github.com/theopenlane/core/v2/pkg/middleware/transaction"
)

func TestAfterCommitRunsAfterCommit(t *testing.T) {
	s := suite.setupTest(t)

	current, err := s.client.Tx(context.Background())
	assert.NilError(t, err)

	ctx := generated.NewTxContext(context.Background(), current)

	var calls int

	enttx.AfterCommit(ctx, func() { calls++ })
	assert.Check(t, is.Equal(0, calls))

	assert.NilError(t, current.Commit())
	assert.Check(t, is.Equal(1, calls))
}

func TestAfterCommitSkippedOnRollback(t *testing.T) {
	s := suite.setupTest(t)

	current, err := s.client.Tx(context.Background())
	assert.NilError(t, err)

	ctx := generated.NewTxContext(context.Background(), current)

	var calls int

	enttx.AfterCommit(ctx, func() { calls++ })

	assert.NilError(t, current.Rollback())
	assert.Check(t, is.Equal(0, calls))
}

func TestTxFromContext(t *testing.T) {
	s := suite.setupTest(t)

	graphTx, err := s.client.Tx(context.Background())
	assert.NilError(t, err)

	restTx, err := s.client.Tx(context.Background())
	assert.NilError(t, err)

	t.Cleanup(func() {
		assert.Check(t, graphTx.Rollback())
		assert.Check(t, restTx.Rollback())
	})

	graphCtx := generated.NewTxContext(context.Background(), graphTx)
	assert.Check(t, enttx.FromContext(graphCtx) == graphTx)

	restCtx := transaction.NewContext(context.Background(), restTx)
	assert.Check(t, enttx.FromContext(restCtx) == restTx)

	assert.Check(t, enttx.FromContext(transaction.NewContext(graphCtx, restTx)) == graphTx)
}
