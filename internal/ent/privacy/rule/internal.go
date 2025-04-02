package rule

import (
	"context"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
)

type contextKey string

const (
	internalCtx contextKey = "internalCtx"
)

func WithInternalContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, internalCtx, true)
}

// IsInternalRequest checks if the context has an internal request key
func IsInternalRequest(ctx context.Context) bool {
	_, ok := ctx.Value(internalCtx).(bool)
	return ok
}

// AllowIfInternalRequest is a pre-policy rule that allows all operations if the context has an internal request key
func AllowIfInternalRequest() privacy.QueryMutationRule {
	return privacy.ContextQueryMutationRule(func(ctx context.Context) error {
		if IsInternalRequest(ctx) {
			return privacy.Allow
		}
		return privacy.Skip
	})
}
