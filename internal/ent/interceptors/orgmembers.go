package interceptors

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/gqlgen-plugins/graphutils"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	fgamodel "github.com/theopenlane/core/fga/model"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/logx"
)

// TraverseOrgMembers is middleware to change the Org Members query
func TraverseOrgMembers() ent.Interceptor {
	return intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
		if auth.IsSystemAdminFromContext(ctx) {
			return nil
		}

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

		caller, ok := auth.CallerFromContext(ctx)
		if !ok || caller == nil {
			logx.FromContext(ctx).Error().Msg("unable to get authenticated user context while traversing org members")
			return auth.ErrNoAuthUser
		}

		orgIDs := caller.OrgIDs()

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

			res, err := dedupeOrgMembers(ctx, members)
			if err != nil {
				return nil, err
			}

			if !graphutils.CheckForRequestedField(ctx, "additionalRoles") {
				return res, nil
			}

			// get additional roles per result
			for i, r := range res {
				res[i].AdditionalRoles, err = getFunctionalRoles(ctx, r.UserID)
				if err != nil {
					return nil, err
				}
			}

			return res, nil
		})
	})
}

func getFunctionalRoles(ctx context.Context, userID string) ([]string, error) {
	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil {
		return []string{}, nil
	}

	roles, err := fgamodel.OrganizationRoles()
	if err != nil {
		return []string{}, err
	}

	ids := make([]string, 0, len(roles))
	for _, role := range roles {
		ids = append(ids, role.ID)
	}

	req := fgax.ListAccess{
		SubjectType: auth.UserSubjectType,
		SubjectID:   userID,
		ObjectID:    caller.OrganizationID,
		ObjectType:  fgax.Kind(generated.TypeOrganization),
		Relations:   ids,
		Context:     utils.NewOrganizationContextKey(caller.SubjectEmail),
	}

	client := utils.AuthzClientFromContext(ctx)
	if client == nil {
		return []string{}, nil
	}

	assignedRoles, err := client.ListRelations(ctx, req)
	if err != nil {
		return []string{}, err
	}

	functionRoles := fgamodel.GetOrganizationRoleStrings(roles, assignedRoles)

	return functionRoles, nil
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
