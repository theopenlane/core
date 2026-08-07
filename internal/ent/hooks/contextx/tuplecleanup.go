package contextx

import (
	"context"

	utilsctx "github.com/theopenlane/utils/contextx"
)

// tupleCleanupKey forces relationship tuple cleanup on delete even for internal requests
var tupleCleanupKey = utilsctx.NewKey[bool]()

// WithTupleCleanup returns a new context that forces the delete permissions hook to run even though
// the request is an internal one. The organization cascade delete runs as an internal caller so it
// can bypass privacy rules, but the records it removes still need their tuples cleaned out of FGA,
// otherwise every cascaded object leaves its relationships behind pointing at rows that are gone
func WithTupleCleanup(ctx context.Context) context.Context {
	return tupleCleanupKey.Set(ctx, true)
}

// TupleCleanupEnabled reports whether tuple cleanup should run despite an internal request
func TupleCleanupEnabled(ctx context.Context) bool {
	cleanup, _ := tupleCleanupKey.Get(ctx)

	return cleanup
}
