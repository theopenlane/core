package tx_test

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/tx"
)

func TestAfterCommitWithoutTransactionRunsImmediately(t *testing.T) {
	var calls int

	tx.AfterCommit(context.Background(), func() { calls++ })

	assert.Check(t, is.Equal(1, calls))
}

func TestFromContextWithoutTransaction(t *testing.T) {
	assert.Check(t, tx.FromContext(context.Background()) == nil)
}
