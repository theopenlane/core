package interceptors

import (
	"context"

	"entgo.io/ent"

	"github.com/datumforge/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/pkg/auth"
)

// InterceptorOrganization is middleware to change the Organization query
func InterceptorOrganization() ent.Interceptor {
	return intercept.TraverseOrganization(func(ctx context.Context, q *generated.OrganizationQuery) error {
		// by pass checks on invite or pre-allowed request
		if _, allow := privacy.DecisionFromContext(ctx); allow {
			return nil
		}

		if rule.ContextHasPrivacyTokenOfType(ctx, &token.OrgInviteToken{}) ||
			rule.ContextHasPrivacyTokenOfType(ctx, &token.SignUpToken{}) {
			return nil
		}

		// if this is an API token, only allow the query if it is for the organization
		if auth.IsAPITokenAuthentication(ctx) {
			orgID, err := auth.GetOrganizationIDFromContext(ctx)
			if err != nil {
				return err
			}

			q.Where(organization.IDEQ(orgID))

			return nil
		}

		if q.EntConfig.Flags.UseListObjectService {
			return filterAllActorOrgIDsFromFGA(ctx, q)
		}

		return filterAllActorOrgIDsFromDB(ctx, q)
	})
}

// filterAllActorOrgIDsFromFGA returns all the organization IDs that the user is a member of, either directly or indirectly via a parent organization
// using the FGA ListObjects service
func filterAllActorOrgIDsFromFGA(ctx context.Context, q *generated.OrganizationQuery) error {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	req := fgax.ListRequest{
		SubjectID:  userID,
		ObjectType: "organization",
	}

	listObjectsResp, err := q.Authz.ListObjectsRequest(ctx, req)
	if err != nil {
		return err
	}

	orgIDs := make([]string, 0, len(listObjectsResp.Objects))

	for _, obj := range listObjectsResp.Objects {
		entity, err := fgax.ParseEntity(obj)
		if err != nil {
			return err
		}

		orgIDs = append(orgIDs, entity.Identifier)
	}

	q.Where(organization.IDIn(orgIDs...))

	return nil
}

// filterAllActorOrgIDsFromDB returns all the organization IDs that the user is a member of, either directly or indirectly via a parent organization
// by querying the database directly and recursively getting all the child organizations
func filterAllActorOrgIDsFromDB(ctx context.Context, q *generated.OrganizationQuery) error {
	// allow the request, otherwise we would be in an infinite loop, as this function is called by the interceptor
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	directOrgIDs, err := generated.FromContext(allowCtx).Organization.
		Query().
		Where(organization.HasMembersWith(orgmembership.HasUserWith(user.ID(userID)))).
		Select(organization.FieldID).
		Strings(allowCtx)
	if err != nil {
		return err
	}

	orgIDs, err := getAllRelatedChildOrgs(allowCtx, directOrgIDs)
	if err != nil {
		return err
	}

	q.Where(organization.IDIn(orgIDs...))

	return nil
}

// getAllRelatedChildOrgs returns all the combined directly related orgs in addition to any child organizations of the parent organizations
func getAllRelatedChildOrgs(ctx context.Context, directOrgIDs []string) ([]string, error) {
	allOrgsIDs := directOrgIDs

	for _, id := range directOrgIDs {
		co, err := getChildOrgIDs(ctx, id)
		if err != nil {
			return nil, err
		}

		allOrgsIDs = append(allOrgsIDs, co...)
	}

	return allOrgsIDs, nil
}

// getChildOrgIDs returns all the child organizations of the parent organization
func getChildOrgIDs(ctx context.Context, parentOrgID string) ([]string, error) {
	// allow the request, otherwise we would be in an infinite loop, as this function is called by the interceptor
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	// given an organization id, get all the children of that organization
	childOrgs, err := generated.FromContext(ctx).Organization.
		Query().
		Where(
			organization.HasParentWith(organization.ID(parentOrgID)),
		).
		Select(organization.FieldID).
		Strings(allowCtx)
	if err != nil {
		return nil, err
	}

	allOrgsIDs := childOrgs

	for _, orgID := range childOrgs {
		// recursively get all the children of the children
		coIDs, err := getChildOrgIDs(ctx, orgID)
		if err != nil {
			return nil, err
		}

		allOrgsIDs = append(allOrgsIDs, coIDs...)
	}

	return allOrgsIDs, nil
}

// getAllParentOrgIDs returns all the parent organization IDs of the child organizations
func getAllParentOrgIDs(ctx context.Context, childOrgIDs []string) ([]string, error) {
	allOrgsIDs := childOrgIDs

	for _, id := range childOrgIDs {
		co, err := getParentOrgIDs(ctx, id)
		if err != nil {
			return nil, err
		}

		allOrgsIDs = append(allOrgsIDs, co...)
	}

	return allOrgsIDs, nil
}

// getParentOrgIDs returns all the parent organizations of the child organization
// this should only be used to get the org members for the current org
// and does not imply the current user is a member or has access to the parent orgs
func getParentOrgIDs(ctx context.Context, childOrgID string) ([]string, error) {
	// allow the request, otherwise we would be in an infinite loop, as this function is called by the interceptor
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	parentOrgs, err := generated.FromContext(ctx).Organization.
		Query().
		Where(
			organization.HasChildrenWith(organization.ID(childOrgID)),
		).
		Select(organization.FieldID).
		Strings(allowCtx)
	if err != nil {
		return nil, err
	}

	allOrgsIDs := parentOrgs

	for _, orgID := range parentOrgs {
		// recursively get all the children of the children
		coIDs, err := getParentOrgIDs(ctx, orgID)
		if err != nil {
			return nil, err
		}

		allOrgsIDs = append(allOrgsIDs, coIDs...)
	}

	return allOrgsIDs, nil
}
