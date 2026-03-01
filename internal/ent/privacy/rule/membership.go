package rule

import (
	"context"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
)

// AllowSelfOrgMembershipDelete allows users to delete only their own org membership
func AllowSelfOrgMembershipDelete() privacy.OrgMembershipMutationRuleFunc {
	return privacy.OrgMembershipMutationRuleFunc(func(ctx context.Context, m *generated.OrgMembershipMutation) error {
		caller, ok := auth.CallerFromContext(ctx)
		if !ok || caller == nil || caller.SubjectID == "" {
			return privacy.Skipf("unable to get user ID from context")
		}

		userID := caller.SubjectID

		id, ok := m.ID()
		if !ok {
			return privacy.Skip
		}

		orgMembership, err := m.Client().OrgMembership.Get(ctx, id)
		if err != nil {
			return privacy.Skipf("unable to get org membership: %v", err)
		}

		if orgMembership.UserID == userID {
			return privacy.Allow
		}

		return privacy.Skipf("user can only delete their own membership")
	})
}
