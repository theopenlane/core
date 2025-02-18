package rule

import (
	"context"
	"errors"

	"entgo.io/ent"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
)

// CheckCurrentOrgAccess checks if the authenticated user has access to the organization
// based on the relation provided
// This rule assumes that the organization id and user id are set in the context
// and only checks for access to the single organization
func CheckCurrentOrgAccess(ctx context.Context, relation string) error {
	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return err
	}

	return checkOrgAccess(ctx, relation, orgID)
}

// CheckOrgAccessBasedOnRequest checks if the authenticated user has access to the organizations that are requested
// in the organization query based on the relation provided
func CheckOrgAccessBasedOnRequest(ctx context.Context, relation string, query *generated.OrganizationQuery) error {
	// run the query with allow context to get the list of organizations
	// the user is trying to access
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	requestedOrgs, err := query.Clone().Select("id").All(allowCtx)
	if err != nil {
		return err
	}

	if len(requestedOrgs) == 0 {
		// return nil if no organizations were found
		// to allow the next check to run
		return nil
	}

	for _, org := range requestedOrgs {
		if err := checkOrgAccess(ctx, relation, org.ID); err != nil && errors.Is(err, privacy.Deny) {
			return err
		}
	}

	return privacy.Allow
}

// checkOrgAccess checks if the authenticated user has access to the organization
func checkOrgAccess(ctx context.Context, relation, organizationID string) error {
	// skip if permission is already set to allow
	if _, allow := privacy.DecisionFromContext(ctx); allow {
		return nil
	}

	log.Debug().Str("relation", relation).Msg("checking access to organization")

	au, err := auth.GetAuthenticatedUserContext(ctx)
	if err != nil {
		return err
	}

	ac := fgax.AccessCheck{
		SubjectID:   au.SubjectID,
		SubjectType: auth.GetAuthzSubjectType(ctx),
		Relation:    relation,
		ObjectType:  generated.TypeOrganization,
		ObjectID:    organizationID,
		Context:     utils.NewOrganizationContextKey(au.SubjectEmail),
	}

	access, err := utils.AuthzClientFromContext(ctx).CheckOrgAccess(ctx, ac)
	if err != nil {
		return err
	}

	if access {
		log.Debug().Str("relation", relation).Msg("access allowed for organization")

		return privacy.Allow
	}

	// deny if it was a mutation is not allowed
	return privacy.Deny
}

// HasOrgMutationAccess is a rule that returns allow decision if user has edit or delete access
func HasOrgMutationAccess() privacy.OrganizationMutationRuleFunc {
	return privacy.OrganizationMutationRuleFunc(func(ctx context.Context, m *generated.OrganizationMutation) error {
		log.Debug().Msg("checking mutation access")

		relation := fgax.CanEdit
		if m.Op().Is(ent.OpDelete | ent.OpDeleteOne) {
			relation = fgax.CanDelete
		}

		user, err := auth.GetAuthenticatedUserContext(ctx)
		if err != nil {
			return err
		}

		ac := fgax.AccessCheck{
			SubjectID:   user.SubjectID,
			SubjectType: auth.GetAuthzSubjectType(ctx),
			Relation:    relation,
			Context:     utils.NewOrganizationContextKey(user.SubjectEmail),
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
					log.Debug().Str("relation", relation).Str("organization_id", parentOrgID).
						Msg("access denied to parent org")

					return generated.ErrPermissionDenied
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
		return generated.ErrPermissionDenied
	})
}
