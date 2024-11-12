package rule

import (
	"context"

	"entgo.io/ent"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
)

// HasOrgMutationAccess is a rule that returns allow decision if user has edit or delete access
func HasOrgMutationAccess() privacy.OrganizationMutationRuleFunc {
	return privacy.OrganizationMutationRuleFunc(func(ctx context.Context, m *generated.OrganizationMutation) error {
		log.Debug().Msg("checking mutation access")

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
					log.Debug().Str("relation", relation).
						Str("organization_id", parentOrgID).
						Msg("access denied to parent org")

					return privacy.Deny
				}
			}

			return privacy.Skip
		}

		// check the organization from the mutation
		oID, _ := m.ID()

		// if it's not set return an error
		if oID == "" {
			log.Debug().Msg("missing expected organization id")

			return privacy.Denyf("missing organization ID information in context")
		}

		log.Info().Str("relation", relation).
			Str("organization_id", oID).
			Msg("checking relationship tuples")

		// check access to the organization
		ac.ObjectID = oID

		access, err := m.Authz.CheckOrgAccess(ctx, ac)
		if err != nil {
			return privacy.Skipf("unable to check access, %s", err.Error())
		}

		if access {
			log.Debug().Str("relation", relation).
				Str("organization_id", oID).
				Msg("access allowed")

			return privacy.Allow
		}

		// deny if it was a mutation is not allowed
		return privacy.Deny
	})
}

// CanCreateObjectsInOrg is a rule that returns allow decision if user has edit access in the organization
// which allows them to create organization owned objects
// This rule is used for objects that are owned by an organization but also offer their
// own permission sets (e.g. groups and programs)
func CanCreateObjectsInOrg() privacy.MutationRuleFunc {
	return privacy.MutationRuleFunc(func(ctx context.Context, m generated.Mutation) error {
		oID, err := getOwnerIDFromEntMutation(m)
		if err != nil || oID == "" {
			// get organization from the auth context
			var err error

			oID, err = auth.GetOrganizationIDFromContext(ctx)
			if err != nil || oID == "" {
				return privacy.Skipf("no owner set on request, cannot check access")
			}
		}

		log.Debug().Msg("checking mutation access")

		relation := fgax.CanEdit
		if m.Op().Is(ent.OpDelete | ent.OpDeleteOne) {
			relation = fgax.CanDelete
		}

		userID, err := auth.GetUserIDFromContext(ctx)
		if err != nil {
			return err
		}

		log.Info().Str("relation", relation).
			Str("organization_id", oID).
			Msg("checking relationship tuples")

		ac := fgax.AccessCheck{
			SubjectID:   userID,
			SubjectType: auth.GetAuthzSubjectType(ctx),
			ObjectID:    oID,
			Relation:    relation,
		}

		access, err := generated.FromContext(ctx).Authz.CheckOrgAccess(ctx, ac)
		if err != nil {
			return privacy.Skipf("unable to check access, %s", err.Error())
		}

		if access {
			log.Debug().Str("relation", relation).
				Str("organization_id", oID).
				Msg("access allowed")

			return privacy.Allow
		}

		// deny if it was a mutation is not allowed
		return privacy.Deny
	})
}

// getOwnerIDFromEntMutation extracts the object id from a the mutation
// by attempting to cast the mutation to a group or program mutation
// if additional object types are needed, they should be added to this function
func getOwnerIDFromEntMutation(m generated.Mutation) (string, error) {
	if o, ok := m.(*generated.GroupMutation); ok {
		if ownerID, ok := o.OwnerID(); ok {
			return ownerID, nil
		}

		return "", nil
	}

	if o, ok := m.(*generated.ProgramMutation); ok {
		if ownerID, ok := o.OwnerID(); ok {
			return ownerID, nil
		}

		return "", nil
	}

	return "", nil
}
