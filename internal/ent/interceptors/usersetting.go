package interceptors

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/internal/ent/generated/usersetting"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
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

		caller, ok := auth.CallerFromContext(ctx)

		switch qCtx.Type {
		// if we are looking at a user in the context of an organization or group
		// filter for just those users
		case "OrgMembership", "GroupMembership":
			if ok && caller != nil {
				q.Where(usersetting.HasUserWith(
					user.HasOrgMembershipsWith(
						orgmembership.HasOrganizationWith(
							organization.IDIn(caller.OrgIDs()...),
						),
					)),
				)

				return nil
			}
		default:
			// if we are looking at self
			if ok && caller != nil && caller.SubjectID != "" {
				q.Where(usersetting.UserID(caller.SubjectID))

				return nil
			}
		}

		return nil
	})
}
