package interceptors

import (
	"context"
	"strings"

	"entgo.io/ent"
	"github.com/99designs/gqlgen/graphql"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
)

// TraverseUser returns an ent interceptor for user that filters users based on the context of the query
func TraverseUser() ent.Interceptor {
	return intercept.TraverseUser(func(ctx context.Context, q *generated.UserQuery) error {
		// bypass filter if the request is allowed, this happens when a user is
		// being created, via invite or other method by another authenticated user
		// or in tests
		if _, allow := privacy.DecisionFromContext(ctx); allow || rule.IsInternalRequest(ctx) {
			return nil
		}

		// allow system admins to see all users
		if auth.IsSystemAdminFromContext(ctx) {
			return nil
		}

		// allow users to be created without filtering
		rootFieldCtx := graphql.GetRootFieldContext(ctx)
		if rootFieldCtx != nil && rootFieldCtx.Object == "createUser" {
			return nil
		}

		switch userFilterType(ctx) {
		// if we are looking at a user in the context of an organization or group
		// filter for just those users
		case "org":
			return filterUsingFGA(ctx, q)
		case "user":
			// if we are looking at self
			if caller, ok := auth.CallerFromContext(ctx); ok && caller != nil && caller.SubjectID != "" {
				q.Where(user.ID(caller.SubjectID))

				return nil
			}
		default:
			// if we want to get all users, don't apply any filters
			return nil
		}

		return nil
	})
}

const (
	// userFilterTypeOrg is the filter type for organization level filtering
	userFilterTypeOrg = "org"
	// userFilterTypeUser is the filter type for user level filtering
	userFilterTypeUser = "user"
	// userFilterTypeNone is the filter type for no filtering
	userFilterTypeNone = ""
)

// userFilterType returns the type of filter to apply to the query
// when querying for users. This is based on the context of the query
// if the root field being requested is a `user` it will filter on the authorized user
// generally, requests will be filtered on the organization, to be able to see all users
// within the organization
func userFilterType(ctx context.Context) string {
	rootFieldCtx := graphql.GetRootFieldContext(ctx)

	if rootFieldCtx != nil && rootFieldCtx.Object != "" {
		nonOrgFilter := []string{
			"user",
		}

		orgFilter := true

		for _, t := range nonOrgFilter {
			if strings.Contains(strings.ToLower(rootFieldCtx.Object), t) {
				orgFilter = false
				break
			}
		}

		if orgFilter {
			return userFilterTypeOrg
		}
	}

	qCtx := ent.QueryFromContext(ctx)
	if qCtx == nil {
		return userFilterTypeNone
	}

	switch qCtx.Type {
	case "Organization":
		return userFilterTypeNone // no filter because this is filtered at the org level, which may be more than one organization
	case "User", "UserSetting":
		return userFilterTypeUser
	default:
		return userFilterTypeOrg
	}
}

// filterUsingFGA filters the user query using the FGA service to get the users with access to the org
func filterUsingFGA(ctx context.Context, q *generated.UserQuery) error {
	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil {
		return auth.ErrNoAuthUser
	}

	if caller.Has(auth.CapBypassFGA) {
		return nil
	}

	orgIDs := caller.OrgIDs()

	userIDs := []string{}

	for _, orgID := range orgIDs {
		req := fgax.ListRequest{
			ObjectID:         orgID,
			ObjectType:       generated.TypeOrganization,
			ConditionContext: utils.NewOrganizationContextKey(""), // use an empty domain context on list
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
