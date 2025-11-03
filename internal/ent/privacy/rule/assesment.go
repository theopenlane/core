package rule

import (
	"context"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/assessment"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
)

// AllowIfAssessmentCreatedBy allows users to edit assessments they created
func AllowIfAssessmentCreatedBy() privacy.AssessmentMutationRuleFunc {
	return privacy.AssessmentMutationRuleFunc(func(ctx context.Context, m *generated.AssessmentMutation) error {
		if m.Op() == generated.OpCreate {
			return privacy.Skip
		}

		userID, err := auth.GetSubjectIDFromContext(ctx)
		if err != nil {
			return privacy.Skipf("unable to get user ID from context: %v", err)
		}

		id, ok := m.ID()
		if !ok {
			return privacy.Skip
		}

		assessment, err := m.Client().Assessment.Get(ctx, id)
		if err != nil {
			return privacy.Skipf("unable to get assessment: %v", err)
		}

		if assessment.CreatedBy == userID {
			return privacy.Allow
		}

		return privacy.Skipf("user is not the creator of this assessment")
	})
}

// AllowIfAssessmentQueryCreatedBy allows users to query only assessments they created
func AllowIfAssessmentQueryCreatedBy() privacy.AssessmentQueryRuleFunc {
	return privacy.AssessmentQueryRuleFunc(func(ctx context.Context, q *generated.AssessmentQuery) error {
		userID, err := auth.GetSubjectIDFromContext(ctx)
		if err != nil {
			return privacy.Skipf("unable to get user ID from context: %v", err)
		}

		q.Where(assessment.CreatedBy(userID))

		return privacy.Allow
	})
}

// AllowIfAssessmentResponseOwner allows users to edit assessment responses they own (where user_id matches)
func AllowIfAssessmentResponseOwner() privacy.AssessmentResponseMutationRuleFunc {
	return privacy.AssessmentResponseMutationRuleFunc(func(ctx context.Context, m *generated.AssessmentResponseMutation) error {
		return privacy.Deny
	})
}

// AllowIfAssessmentResponseQueryOwner allows users to query only assessment responses they own
func AllowIfAssessmentResponseQueryOwner() privacy.AssessmentResponseQueryRuleFunc {
	return privacy.AssessmentResponseQueryRuleFunc(func(ctx context.Context, q *generated.AssessmentResponseQuery) error {
		return privacy.Deny
	})
}
