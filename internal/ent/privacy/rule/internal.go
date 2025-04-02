package rule

import (
	"context"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
)

// IsInternalRequest checks if the context has an internal request key
func IsInternalRequest(ctx context.Context) bool {
	_, ok := ctx.Value(InternalRequestContextKey{}).(bool)
	return ok
}

// AllowInternalRequestRule allows the query/mutation to proceed if the context has an internal request key
func AllowInternalRequestRule() privacy.QueryMutationRule {
	return privacy.ContextQueryMutationRule(func(ctx context.Context) error {
		if IsInternalRequest(ctx) {
			return privacy.Allow
		}

		return privacy.Skipf("no internal request key found in context")
	})
}
