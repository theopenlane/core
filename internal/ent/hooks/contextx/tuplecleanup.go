package contextx

import "context"

// TupleCleanupKey is the context key used to force relationship tuple cleanup on delete
type TupleCleanupKey string

const (
	// TupleCleanup is the context value that forces tuple cleanup even for internal requests
	TupleCleanup TupleCleanupKey = "cascade_delete_tuple_cleanup"
)

// WithTupleCleanup returns a new context that forces the delete permissions hook to run even though
// the request is an internal one. The organization cascade delete runs as an internal caller so it
// can bypass privacy rules, but the records it removes still need their tuples cleaned out of FGA
func WithTupleCleanup(ctx context.Context) context.Context {
	return context.WithValue(ctx, TupleCleanup, true)
}

// TupleCleanupEnabled reports whether tuple cleanup should run despite an internal request
func TupleCleanupEnabled(ctx context.Context) bool {
	cleanup, ok := ctx.Value(TupleCleanup).(bool)

	return ok && cleanup
}
