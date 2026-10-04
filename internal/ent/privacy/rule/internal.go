package rule

import (
	"context"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
)

// AllowIfInternalReadRequest is a query pre-policy rule that allows reads if the caller carries internal operation or internal read capability
func AllowIfInternalReadRequest() privacy.QueryRule {
	return privacy.QueryRuleFunc(func(ctx context.Context, _ generated.Query) error {
		if auth.IsInternalReadRequest(ctx) {
			return privacy.Allow
		}

		return privacy.Skip
	})
}

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
