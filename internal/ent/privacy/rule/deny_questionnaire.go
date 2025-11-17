package rule

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
)

// DenyIfMissingQuestionnaireContext denies a mutation if the context does not have a QuestionnaireContextKey
// This enforces that assessment responses can ONLY be created with a questionnaire JWT (anonymous users)
func DenyIfMissingQuestionnaireContext() privacy.MutationRule {
	return privacy.MutationRuleFunc(func(ctx context.Context, m ent.Mutation) error {

		_, ok := privacy.DecisionFromContext(ctx)
		if ok || IsInternalRequest(ctx) {
			return privacy.Skip
		}

		if m.Op() == ent.OpCreate || m.Op() == ent.OpDeleteOne {
			return privacy.Skip
		}

		if _, ok := contextx.From[auth.QuestionnaireContextKey](ctx); ok {
			return privacy.Skip
		}

		return privacy.Denyf("assessment responses can only be created with a questionnaire context")
	})
}
