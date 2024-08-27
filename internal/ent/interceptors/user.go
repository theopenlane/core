package interceptors

import (
	"context"
	"slices"

	"entgo.io/ent"

	"github.com/99designs/gqlgen/graphql"

	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/pkg/auth"
)

// TraverseUser returns an ent interceptor for user that filters users based on the context of the query
func TraverseUser() ent.Interceptor {
	return intercept.TraverseUser(func(ctx context.Context, q *generated.UserQuery) error {
		// bypass filter if the request is allowed, this happens when a user is
		// being created, via invite or other method by another authenticated user
		// or in tests
		if _, allow := privacy.DecisionFromContext(ctx); allow {
			return nil
		}

		// allow users to be created without filtering
		rootFieldCtx := graphql.GetRootFieldContext(ctx)
		if rootFieldCtx != nil && rootFieldCtx.Object == "createUser" {
			return nil
		}

		switch filterType(ctx) {
		// if we are looking at a user in the context of an organization or group
		// filter for just those users
		case "org":
			if q.EntConfig.Flags.UseListUserService {
				q.Logger.Debug("using FGA to filter users")
				return filterUsingFGA(ctx, q)
			}

			q.Logger.Debug("using the db to filter users")

			return filterUsingDB(ctx, q)
		case "user":
			// if we are looking at self
			userID, err := auth.GetUserIDFromContext(ctx)
			if err == nil {
				q.Where(user.ID(userID))

				return nil
			}
		default:
			// if we want to get all users, don't apply any filters
			return nil
		}

		return nil
	})
}

// filterType returns the type of filter to apply to the query
func filterType(ctx context.Context) string {
	rootFieldCtx := graphql.GetRootFieldContext(ctx)

	// the extended resolvers allow members to be adding on creation or update of a group
	// so we need to filter for the org
	if rootFieldCtx != nil {
		allowedCtx := []string{
			"createGroup",
			"updateGroup",
			"createGroupMembership",
			"updateGroupMembership",
			"organization",
		}

		if slices.Contains(allowedCtx, rootFieldCtx.Object) {
			return "org"
		}
	}

	qCtx := ent.QueryFromContext(ctx)
	if qCtx == nil {
		return ""
	}

	switch qCtx.Type {
	case "OrgMembership", "GroupMembership", "Group":
		return "org"
	case "Organization":
		return "" // no filter because this is filtered at the org level
	default:
		return "user"
	}
}

// filterUsingFGA filters the user query using the FGA service to get the users with access to the org
func filterUsingFGA(ctx context.Context, q *generated.UserQuery) error {
	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return err
	}

	req := fgax.ListRequest{
		ObjectID:   orgID,
		ObjectType: "organization",
	}

	listUserResp, err := q.Authz.ListUserRequest(ctx, req)
	if err != nil {
		return err
	}

	userIDs := []string{}

	for _, user := range listUserResp.Users {
		userIDs = append(userIDs, user.Object.Id)
	}

	q.Where(user.IDIn(userIDs...))

	return nil
}

// filterUsingDB filters the user query using the database to get the the users that are members of the org
func filterUsingDB(ctx context.Context, q *generated.UserQuery) error {
	orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
	if err != nil {
		return err
	}

	// get child orgs in addition to the orgs the user is a direct member of
	childOrgs, err := getAllRelatedChildOrgs(ctx, orgIDs)
	if err != nil {
		return err
	}

	// get the parent orgs of the orgs the user is a member of to ensure all
	// users in the org tree are returned
	allOrgsIDs, err := getAllParentOrgIDs(ctx, orgIDs)
	if err != nil {
		return err
	}

	allOrgsIDs = append(allOrgsIDs, childOrgs...)

	q.Where(user.HasOrgMembershipsWith(
		orgmembership.HasOrganizationWith(
			organization.IDIn(allOrgsIDs...),
		),
	))

	return nil
}
