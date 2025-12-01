package interceptors

import (
	"context"
	"strings"

	"entgo.io/ent"

	"github.com/99designs/gqlgen/graphql"
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

		if ok := rule.IsInternalRequest(ctx); ok {
			return nil
		}

		if rule.ContextHasPrivacyTokenOfType[*token.OrgInviteToken](ctx) ||
			rule.ContextHasPrivacyTokenOfType[*token.SignUpToken](ctx) {
			return nil
		}

		// Authenticated users should have their organizations IDs set in their auth context
		// after logging in, check this first before using the AddIDPredicate and requiring a
		// query to fga
		au, err := auth.GetAuthenticatedUserFromContext(ctx)
		if err == nil && len(au.OrganizationIDs) > 0 {
			// if the request is not using a JWT, we can restrict to the authorized orgs
			// from the context
			if au.AuthenticationType != auth.JWTAuthentication {
				q.WhereP(organization.IDIn(au.OrganizationIDs...))

				return nil
			}

			// if the request is using a JWT and is not org owned, for example user profile, personal access tokens, etc,
			// as well as a query on all organizations for a user,
			// we need to restrict on all organization instead of just the current one
			// we do pattern matching on these types so that things like `deleteOrganization` are included
			// along with a query type of organizations, as an example
			useListObjectsTypes := []string{
				"personalAccessToken", // ability to add multiple orgs to a PAT
				"organization",        // ability to list all orgs user has access to
				"orgMemberships",      // due to parent org relationships, we need to check all orgs
				"userSetting",         // default org ID
				"user",                // default org ID
			}

			// fall back to ListObjects if we are in one of the cases above
			fCtx := graphql.GetFieldContext(ctx)
			fieldCheck := ""

			if fCtx != nil {
				if fCtx.Object == "Query" || fCtx.Object == "Mutation" {
					fieldCheck = fCtx.Field.Name
				} else {
					fieldCheck = fCtx.Object
				}

				if fieldCheck != "" {
					for _, t := range useListObjectsTypes {
						if strings.Contains(strings.ToLower(fieldCheck), strings.ToLower(t)) {
							return AddIDPredicate(ctx, q)
						}
					}
				}
			}

			// other requests can fall back to the authorized orgs
			q.WhereP(organization.IDIn(au.OrganizationIDs...))

			return nil
		}

		// fallback to the AddIDPredicate if we don't have any org IDs in the context
		// this shouldn't happen in normal operation, but is a safety net
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
