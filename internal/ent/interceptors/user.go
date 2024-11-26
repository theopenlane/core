package interceptors

import (
	"context"
	"slices"

	"entgo.io/ent"
	"github.com/99designs/gqlgen/graphql"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/user"
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
			return filterUsingFGA(ctx, q)
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
			"createProgramMembership",
			"updateProgramMembership",
			"createProgram",
			"updateProgram",
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
	case "OrgMembership", "GroupMembership", "Group", "ProgramMembership", "Program":
		return "org"
	case "Organization":
		return "" // no filter because this is filtered at the org level
	default:
		return "user"
	}
}

// filterUsingFGA filters the user query using the FGA service to get the users with access to the org
func filterUsingFGA(ctx context.Context, q *generated.UserQuery) error {
	orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
	if err != nil {
		return err
	}

	userIDs := []string{}

	for _, orgID := range orgIDs {
		req := fgax.ListRequest{
			ObjectID:   orgID,
			ObjectType: "organization",
		}

		listUserResp, err := q.Authz.ListUserRequest(ctx, req)
		if err != nil {
			return err
		}

		for _, user := range listUserResp.Users {
			userIDs = append(userIDs, user.Object.Id)
		}
	}

	q.Where(user.IDIn(userIDs...))

	return nil
}
