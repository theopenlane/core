package rule

import (
	"context"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
)

// AllowIfInternalRequest is a pre-policy rule that allows all operations if
// the caller carries internal request capability.
func AllowIfInternalRequest() privacy.QueryMutationRule {
	return privacy.ContextQueryMutationRule(func(ctx context.Context) error {
		if auth.IsInternalRequest(ctx) {
			return privacy.Allow
		}

		return privacy.Skip
	})
}
