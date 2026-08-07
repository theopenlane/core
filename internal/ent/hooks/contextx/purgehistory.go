package contextx

import (
	"context"

	utilsctx "github.com/theopenlane/utils/contextx"
)

// purgeHistoryKey marks a cascade delete as purging history rather than recording it
var purgeHistoryKey = utilsctx.NewKey[bool]()

// WithPurgeHistory returns a new context that opts the cascade delete into purging history rows.
// EdgeCleanup deletes the history for each record it removes only when this is set, so callers that
// want to keep the audit trail are unaffected
func WithPurgeHistory(ctx context.Context) context.Context {
	return purgeHistoryKey.Set(ctx, true)
}

// PurgeHistoryEnabled reports whether the cascade delete should purge history rows
func PurgeHistoryEnabled(ctx context.Context) bool {
	purge, _ := purgeHistoryKey.Get(ctx)

	return purge
}
