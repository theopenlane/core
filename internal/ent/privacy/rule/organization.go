package rule

import (
	"context"
	"errors"
	"slices"

	"entgo.io/ent"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/permissioncache"
)

// ownerMutation is a minimal mutation contract used to resolve organization ownership.
type ownerMutation interface {
	OwnerID() (id string, exists bool)
}

// CheckCurrentOrgAccess checks if the authenticated user has access to the organization
// based on the relation provided
// This rule assumes that the organization id and user id are set in the context
// and only checks for access to the single organization
func CheckCurrentOrgAccess(ctx context.Context, m ent.Mutation, relation string) error {
	logx.FromContext(ctx).Debug().Str("relation", relation).Msg("checking access for organization")
	// skip if permission is already set to allow or if it's an internal request
	if _, allow := privacy.DecisionFromContext(ctx); allow || IsInternalRequest(ctx) {
		return privacy.Allow
	}

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err == nil {
		if relation == fgax.CanView {
			// if the relation is view, we can skip the check
			return privacy.Allow
		}

		return checkOrgAccess(ctx, relation, orgID)
	}

	// else we need to get the object id from the mutation and get the owner id, this should only happen on deletes when using personal access tokens
	mut, ok := m.(ownerMutation)
	if ok {
		orgID, ok = mut.OwnerID()
		if ok && orgID != "" {
			return checkOrgAccess(ctx, relation, orgID)
		}
	}

	// allow it to continue, there is a hook that will filter on the user's authorized org ids
	// see mixin_orgowned.go:defaultOrgHookFunc() for more details
	return privacy.Allow
}

// CheckOrgAccessBasedOnRequest checks if the authenticated user has access to the organizations that are requested
// in the organization query based on the relation provided
func CheckOrgAccessBasedOnRequest(ctx context.Context, relation string, query *generated.OrganizationQuery) error {
	// skip if it's an internal request
	if IsInternalRequest(ctx) {
		return privacy.Allow
	}

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
// and logs additional context about the mutation if provided
func checkOrgAccess(ctx context.Context, relation, organizationID string) error {
	// skip if permission is already set to allow or if it's an internal request
	if _, allow := privacy.DecisionFromContext(ctx); allow || IsInternalRequest(ctx) {
		return nil
	}

	au, err := auth.GetAuthenticatedUserFromContext(ctx)
	if err != nil {
		return err
	}

	if ok := orgAccessBasedOnContext(ctx, relation, au.OrganizationRole); ok {
		return privacy.Allow
	}

	if slices.Contains(au.OrganizationIDs, organizationID) && relation == fgax.CanView {
		logx.FromContext(ctx).Debug().Str("relation", relation).Msg("access allowed for organization based on user's orgs")

		return privacy.Allow
	}

	// check the cache first
	if cache, ok := permissioncache.CacheFromContext(ctx); ok {
		if hasRole, err := cache.HasRole(ctx, au.SubjectID, organizationID, relation); err == nil && hasRole {
			logx.FromContext(ctx).Debug().Str("relation", relation).Msg("access allowed for organization based on cache")

			return privacy.Allow
		}
	}

	ac := fgax.AccessCheck{
		SubjectID:   au.SubjectID,
		SubjectType: auth.GetAuthzSubjectType(ctx),
		Relation:    relation,
		ObjectID:    organizationID,
		Context:     utils.NewOrganizationContextKey(au.SubjectEmail),
	}

	access, err := utils.AuthzClientFromContext(ctx).CheckOrgAccess(ctx, ac)
	if err != nil {
		return err
	}

	if access {
		logx.FromContext(ctx).Debug().Str("relation", relation).Msg("access allowed for organization based on fga")

		if cache, ok := permissioncache.CacheFromContext(ctx); ok {
			if err := cache.SetRole(ctx, au.SubjectID, organizationID, relation); err != nil {
				logx.FromContext(ctx).Err(err).Msg("failed to set role cache")
			}
		}

		return privacy.Allow
	}

	// deny if it was a mutation is not allowed
	// we check owner relation to skip group level checks, but this ends up being a deny for non-owners
	// and creates noise in the logs; we want to to log at debug level when the check is owners only and info otherwise
	if relation == fgax.OwnerRelation {
		logx.FromContext(ctx).Debug().Str("relation", relation).Str("subject_id", au.SubjectID).Str("email", au.SubjectEmail).Str("organization_id", organizationID).Str("auth_type", string(au.AuthenticationType)).Msg("request denied by access for user in organization")
	} else {
		logx.FromContext(ctx).Info().Str("relation", relation).Str("subject_id", au.SubjectID).Str("email", au.SubjectEmail).Str("organization_id", organizationID).Str("auth_type", string(au.AuthenticationType)).Msg("request denied by ownership access for user in organization")
	}

	return generated.ErrPermissionDenied
}

// orgAccessBasedOnContext checks if the user has access to the organization based on their role in the organization
func orgAccessBasedOnContext(ctx context.Context, relation string, orgRole auth.OrganizationRoleType) bool {
	// skip early if we have the role of owner
	switch relation {
	case fgax.OwnerRelation, fgax.CanDelete:
		// owners have the owner relation and have delete access
		if orgRole == auth.OwnerRole {
			return true
		}
	case fgax.CanEdit:
		// super admins and owners have full edit access and do not need check do explicit checks
		if auth.HasFullOrgWriteAccessFromContext(ctx) {
			return true
		}
	}

	return false
}

// HasOrgMutationAccess is a rule that returns allow decision if user has edit or delete access
func HasOrgMutationAccess() privacy.OrganizationMutationRuleFunc {
	return privacy.OrganizationMutationRuleFunc(func(ctx context.Context, m *generated.OrganizationMutation) error {
		logx.FromContext(ctx).Debug().Msg("checking mutation access")

		relation := fgax.CanEdit
		if m.Op().Is(ent.OpDelete | ent.OpDeleteOne) {
			relation = fgax.CanDelete
		}

		user, err := auth.GetAuthenticatedUserFromContext(ctx)
		if err != nil {
			return err
		}

		// check the cache first
		if cache, ok := permissioncache.CacheFromContext(ctx); ok {
			if hasRole, err := cache.HasRole(ctx, user.SubjectID, user.OrganizationID, relation); err == nil && hasRole {
				logx.FromContext(ctx).Debug().Str("relation", relation).Msg("access allowed for organization based on cache")

				return privacy.Allow
			}
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
					logx.FromContext(ctx).Error().Str("relation", relation).Str("entity_type", m.Type()).Str("operation", m.Op().String()).Str("organization_id", parentOrgID).Str("auth_type", string(user.AuthenticationType))

					return generated.ErrPermissionDenied
				}
			}

			return privacy.Skip
		}

		// check the organization from the mutation
		oID, _ := m.ID()

		// if it's not set return an error
		if oID == "" {
			logx.FromContext(ctx).Debug().Msg("missing expected organization id")

			return privacy.Denyf("missing organization ID information in context")
		}

		logx.FromContext(ctx).Debug().Str("relation", relation).
			Str("organization_id", oID).
			Msg("checking relationship tuples")

		// check access to the organization
		ac.ObjectID = oID

		access, err := m.Authz.CheckOrgAccess(ctx, ac)
		if err != nil {
			return privacy.Skipf("unable to check access, %s", err.Error())
		}

		if access {
			logx.FromContext(ctx).Debug().Str("relation", relation).
				Str("organization_id", oID).
				Msg("access allowed")

			if cache, ok := permissioncache.CacheFromContext(ctx); ok {
				if err := cache.SetRole(ctx, user.SubjectID, oID, relation); err != nil {
					logx.FromContext(ctx).Err(err).Msg("failed to set role cache")
				}
			}

			return privacy.Allow
		}

		// deny if it was a mutation is not allowed
		logx.FromContext(ctx).Info().Str("relation", relation).Str("entity_type", m.Type()).Str("operation", m.Op().String()).Str("subject_id", user.SubjectID).Str("email", user.SubjectEmail).Str("organization_id", oID).Str("auth_type", string(user.AuthenticationType))

		return generated.ErrPermissionDenied
	})
}
