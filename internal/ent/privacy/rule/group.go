package rule

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
)

// CanCreateGroupsInOrg is a rule that returns allow decision if user has edit access in the organization
func CanCreateGroupsInOrg() privacy.GroupMutationRuleFunc {
	return privacy.GroupMutationRuleFunc(func(ctx context.Context, m *generated.GroupMutation) error {
		oID, ok := m.OwnerID()
		if !ok || oID == "" {
			// get organization from the auth context
			var err error

			oID, err = auth.GetOrganizationIDFromContext(ctx)
			if err != nil || oID == "" {
				return privacy.Skipf("no owner set on request, cannot check access")
			}
		}

		m.Logger.Debugw("checking mutation access")

		relation := fgax.CanEdit
		if m.Op().Is(ent.OpDelete | ent.OpDeleteOne) {
			relation = fgax.CanDelete
		}

		userID, err := auth.GetUserIDFromContext(ctx)
		if err != nil {
			return err
		}

		m.Logger.Infow("checking relationship tuples", "relation", relation, "organization_id", oID)

		ac := fgax.AccessCheck{
			SubjectID:   userID,
			SubjectType: auth.GetAuthzSubjectType(ctx),
			ObjectID:    oID,
			Relation:    relation,
		}

		access, err := m.Authz.CheckOrgAccess(ctx, ac)
		if err != nil {
			return privacy.Skipf("unable to check access, %s", err.Error())
		}

		if access {
			m.Logger.Debugw("access allowed", "relation", relation, "organization_id", oID)

			return privacy.Allow
		}

		// deny if it was a mutation is not allowed
		return privacy.Deny
	})
}
