package rule

import (
	"context"

	"github.com/theopenlane/ent/generated/privacy"
)

// AllowIfContextAllowRule allows the query to proceed if the context has an allow rule
func AllowIfContextAllowRule() privacy.QueryMutationRule {
	return privacy.ContextQueryMutationRule(func(ctx context.Context) error {
		if _, allow := privacy.DecisionFromContext(ctx); allow {
			return privacy.Allow
		}

		return privacy.Skipf("no allow rule found in context")
	})
}
