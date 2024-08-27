package interceptors

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/internal/ent/generated/usersetting"
	"github.com/theopenlane/core/pkg/auth"
)

func InterceptorUserSetting() ent.Interceptor {
	return intercept.TraverseUserSetting(func(ctx context.Context, q *generated.UserSettingQuery) error {
		// bypass filter if the request is allowed, this happens when a user is
		// being created, via invite or other method by another authenticated user
		// or in tests
		if _, allow := privacy.DecisionFromContext(ctx); allow {
			return nil
		}

		qCtx := ent.QueryFromContext(ctx)
		if qCtx == nil {
			return nil
		}

		switch qCtx.Type {
		// if we are looking at a user in the context of an organization or group
		// filter for just those users
		case "OrgMembership", "GroupMembership":
			orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
			if err == nil {
				q.Where(usersetting.HasUserWith(
					user.HasOrgMembershipsWith(
						orgmembership.HasOrganizationWith(
							organization.IDIn(orgIDs...),
						),
					)),
				)

				return nil
			}
		default:
			// if we are looking at self
			userID, err := auth.GetUserIDFromContext(ctx)
			if err == nil {
				q.Where(usersetting.UserID(userID))

				return nil
			}
		}

		return nil
	})
}
