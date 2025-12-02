package rule

import (
	"context"

	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/ent/generated/privacy"
)

type internalAllowContextKey struct{}

// WithInternalContext adds an internal request key to the context
func WithInternalContext(ctx context.Context) context.Context {
	return contextx.With(ctx, internalAllowContextKey{})
}

// IsInternalRequest checks if the context has an internal request key
func IsInternalRequest(ctx context.Context) bool {
	_, ok := contextx.From[internalAllowContextKey](ctx)
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
