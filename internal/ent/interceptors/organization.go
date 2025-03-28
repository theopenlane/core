package interceptors

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/token"
)

// InterceptorOrganization is middleware to change the Organization query
func InterceptorOrganization() ent.Interceptor {
	return intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
		// by pass checks on invite or pre-allowed request
		if _, allow := privacy.DecisionFromContext(ctx); allow {
			return nil
		}

		if rule.ContextHasPrivacyTokenOfType[*token.OrgInviteToken](ctx) ||
			rule.ContextHasPrivacyTokenOfType[*token.SignUpToken](ctx) {
			return nil
		}

		// if this is an API token, only allow the query if it is for the organization
		if auth.IsAPITokenAuthentication(ctx) {
			orgID, err := auth.GetOrganizationIDFromContext(ctx)
			if err != nil {
				return err
			}

			q.WhereP(organization.IDEQ(orgID))

			return nil
		}

		// use the id predicate; there will never be a large list of orgs
		// so its safe to use list objects ahead of time
		return AddIDPredicate(ctx, q)
	})
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
