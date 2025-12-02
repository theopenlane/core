package interceptors

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/intercept"
	"github.com/theopenlane/ent/generated/orgmembership"
	"github.com/theopenlane/ent/generated/privacy"
)

// TraverseOrgMembers is middleware to change the Org Members query
func TraverseOrgMembers() ent.Interceptor {
	return intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
		// bypass filter if the request is internal and already set to allowed
		if _, allow := privacy.DecisionFromContext(ctx); allow {
			return nil
		}

		// Organization list queries should not be deduped, they
		// will show up under each org they are a member of
		ctxQuery := ent.QueryFromContext(ctx)
		if ctxQuery.Type == generated.TypeOrganization {
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

			if orgMembersSkipInterceptor(ctx) {
				return v, nil
			}

			// deduplicate the org members if the query is for org members
			members, ok := v.([]*generated.OrgMembership)
			if !ok || len(members) == 0 {
				return v, nil
			}

			return dedupeOrgMembers(ctx, members)
		})
	})
}

// dedupeOrgMembers removes duplicate org members from the list
func dedupeOrgMembers(ctx context.Context, members []*generated.OrgMembership) ([]*generated.OrgMembership, error) {
	seen := map[string]*generated.OrgMembership{}
	deduped := []*generated.OrgMembership{}

	for _, om := range members {
		// we dedupe for hierarchical orgs, we should skip this if the org has no parent or children
		org, err := om.Organization(ctx)
		if err != nil {
			return nil, err
		}

		children, err := org.QueryChildren().All(ctx)
		if err != nil {
			return nil, err
		}

		if org.ParentOrganizationID == "" && len(children) == 0 {
			deduped = append(deduped, om)
			seen[om.UserID] = om

			continue
		}

		if _, ok := seen[om.UserID]; !ok {
			deduped = append(deduped, om)
			seen[om.UserID] = om
		}
	}

	return deduped, nil
}

// orgMembersSkipInterceptor includes conditions to skip the org members interceptor
func orgMembersSkipInterceptor(ctx context.Context) bool {
	// bypass filter if the request is internal and already set to allowed
	// this only happens from internal requests
	// and we don't need to dedupe the org members
	if _, allow := privacy.DecisionFromContext(ctx); allow {
		return true
	}

	// Organization list queries should not be deduped, they
	// will show up under each org they are a member of
	ctxQuery := ent.QueryFromContext(ctx)

	return ctxQuery.Type == generated.TypeOrganization
}
