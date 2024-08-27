package rule

import (
	"context"

	"entgo.io/ent"
	"github.com/datumforge/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/auth"
)

// HasOrgMutationAccess is a rule that returns allow decision if user has edit or delete access
func HasOrgMutationAccess() privacy.OrganizationMutationRuleFunc {
	return privacy.OrganizationMutationRuleFunc(func(ctx context.Context, m *generated.OrganizationMutation) error {
		m.Logger.Debugw("checking mutation access")

		relation := fgax.CanEdit
		if m.Op().Is(ent.OpDelete | ent.OpDeleteOne) {
			relation = fgax.CanDelete
		}

		userID, err := auth.GetUserIDFromContext(ctx)
		if err != nil {
			return err
		}

		ac := fgax.AccessCheck{
			SubjectID:   userID,
			SubjectType: auth.GetAuthzSubjectType(ctx),
			Relation:    relation,
		}

		// No permissions checks on creation of org except if this is not a root org
		if m.Op().Is(ent.OpCreate) {
			parentOrgID, ok := m.ParentID()

			if ok {
				// check the parent organization
				ac.ObjectID = parentOrgID

				access, err := m.Authz.CheckOrgAccess(ctx, ac)
				if err != nil {
					return privacy.Skipf("unable to check access, %s", err.Error())
				}

				if !access {
					m.Logger.Debugw("access denied to parent org", "relation", relation, "organization_id", parentOrgID)

					return privacy.Deny
				}
			}

			return privacy.Skip
		}

		// check the organization from the mutation
		oID, _ := m.ID()

		// if it's not set return an error
		if oID == "" {
			m.Logger.Debugw("missing expected organization id")

			return privacy.Denyf("missing organization ID information in context")
		}

		m.Logger.Infow("checking relationship tuples", "relation", relation, "organization_id", oID)

		// check access to the organization
		ac.ObjectID = oID

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
