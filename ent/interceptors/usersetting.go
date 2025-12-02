package interceptors

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/intercept"
	"github.com/theopenlane/ent/generated/organization"
	"github.com/theopenlane/ent/generated/orgmembership"
	"github.com/theopenlane/ent/generated/privacy"
	"github.com/theopenlane/ent/generated/user"
	"github.com/theopenlane/ent/generated/usersetting"
	"github.com/theopenlane/ent/privacy/rule"
)

func InterceptorUserSetting() ent.Interceptor {
	return intercept.TraverseUserSetting(func(ctx context.Context, q *generated.UserSettingQuery) error {
		// bypass filter if the request is allowed or if it's an internal request
		if _, allow := privacy.DecisionFromContext(ctx); allow || rule.IsInternalRequest(ctx) {
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
			userID, err := auth.GetSubjectIDFromContext(ctx)
			if err == nil {
				q.Where(usersetting.UserID(userID))

				return nil
			}
		}

		return nil
	})
}
