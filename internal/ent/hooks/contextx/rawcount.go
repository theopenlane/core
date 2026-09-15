package contextx

import (
	"context"

	utilsctx "github.com/theopenlane/utils/contextx"
)

// RawCount records how many rows a list query returned before the FGA filter dropped any
type RawCount struct {
	n   int
	set bool
}

var rawCountKey = utilsctx.NewKey[*RawCount]()

// WithRawCount returns a context carrying a RawCount the filter interceptor will populate
func WithRawCount(ctx context.Context) (context.Context, *RawCount) {
	rc := &RawCount{}

	return rawCountKey.Set(ctx, rc), rc
}

// SetRawCount records the pre-filter row count when the context is carrying a RawCount
func SetRawCount(ctx context.Context, n int) {
	rc, ok := rawCountKey.Get(ctx)
	if !ok || rc == nil {
		return
	}

	rc.n = n
	rc.set = true
}

// Value returns the recorded count, or fallback when no filter reported one
func (r *RawCount) Value(fallback int) int {
	if r == nil || !r.set {
		return fallback
	}

	return r.n
}
