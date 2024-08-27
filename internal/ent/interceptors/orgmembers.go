package interceptors

import (
	"context"

	"entgo.io/ent"

	"github.com/99designs/gqlgen/graphql"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/iam/auth"
)

// TraverseOrgMembers is middleware to change the Org Members query
func TraverseOrgMembers() ent.Interceptor {
	return intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
		// bypass filter if the request is internal and already set to allowed
		if _, allow := privacy.DecisionFromContext(ctx); allow {
			return nil
		}

		// Organization list queries should not be filtered by organization id
		ctxQuery := ent.QueryFromContext(ctx)
		if ctxQuery.Type == "Organization" {
			return nil
		}

		orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
		if err != nil {
			return err
		}

		// get all parent orgs to ensure we get all OrgMembers in the org tree
		allOrgsIDs, err := getAllParentOrgIDs(ctx, orgIDs)
		if err != nil {
			return err
		}

		// sets the organization id on the query for the current organization
		q.WhereP(orgmembership.OrganizationIDIn(allOrgsIDs...))

		return nil
	})
}

// InterceptorOrgMember is middleware to change the OrgMember query result
func InterceptorOrgMember() ent.Interceptor {
	return ent.InterceptFunc(func(next ent.Querier) ent.Querier {
		return intercept.OrgMembershipFunc(func(ctx context.Context, q *generated.OrgMembershipQuery) (generated.Value, error) {
			// run the query
			v, err := next.Query(ctx, q)
			if err != nil {
				return nil, err
			}

			// deduplicate the org members if the query is for org members
			members, ok := v.([]*generated.OrgMembership)
			if !ok {
				return v, err
			}

			return dedupeOrgMembers(ctx, members)
		})
	})
}

// dedupeOrgMembers removes duplicate org members from the list
func dedupeOrgMembers(ctx context.Context, members []*generated.OrgMembership) ([]*generated.OrgMembership, error) {
	authorizedOrg, err := getQueriedOrg(ctx)
	if err != nil {
		return nil, err
	}

	seen := map[string]*generated.OrgMembership{}
	deduped := []*generated.OrgMembership{}

	for _, om := range members {
		if _, ok := seen[om.UserID]; ok {
			// prefer the direct membership over the indirect one
			if om.OrganizationID == authorizedOrg {
				seen[om.UserID] = om
			}
		} else {
			deduped = append(deduped, om)
			seen[om.UserID] = om
		}
	}

	return deduped, nil
}

// getQueriedOrg gets the organization id from the context or the graphql context
func getQueriedOrg(ctx context.Context) (string, error) {
	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err == nil {
		return orgID, nil
	}

	// when the organization id is not in the context, try to get it from the graphql context
	// for the Organization query
	gtx := graphql.GetFieldContext(ctx)
	if gtx != nil {
		if gtx.Object == "Organization" {
			if gtx.Parent != nil && gtx.Parent.Args != nil {
				orgID, ok := gtx.Parent.Args["id"]
				if ok {
					return orgID.(string), nil
				}
			}
		}
	}

	return "", ErrRetrievingObjects
}
