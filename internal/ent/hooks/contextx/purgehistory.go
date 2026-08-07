package contextx

import "context"

// PurgeHistoryKey is the context key used to opt a cascade delete into purging the
// history rows of every record it removes
type PurgeHistoryKey string

const (
	// PurgeHistory is the context value that triggers the history purge during a cascade delete
	PurgeHistory PurgeHistoryKey = "cascade_delete_purge_history"
)

// WithPurgeHistory returns a new context that opts the cascade delete into purging history rows.
// EdgeCleanup deletes the history for each record it removes only when this is set, so callers that
// want to keep the audit trail are unaffected
func WithPurgeHistory(ctx context.Context) context.Context {
	return context.WithValue(ctx, PurgeHistory, true)
}

// PurgeHistoryEnabled reports whether the cascade delete should purge history rows
func PurgeHistoryEnabled(ctx context.Context) bool {
	purge, ok := ctx.Value(PurgeHistory).(bool)

	return ok && purge
}
