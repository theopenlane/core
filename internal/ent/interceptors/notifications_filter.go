package interceptors

import (
	"context"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/notification"
	"github.com/theopenlane/core/pkg/logx"
)

// NotificationQueryFilter automatically filters notifications based on user context
func NotificationQueryFilter() generated.Interceptor {
	return generated.TraverseFunc(func(ctx context.Context, q generated.Query) error {
		// Only apply to Notification queries
		nq, ok := q.(*generated.NotificationQuery)
		if !ok {
			return nil
		}

		// Get user info from context
		caller, ok := auth.CallerFromContext(ctx)
		if !ok || caller == nil {
			logx.FromContext(ctx).Error().Msg("unable to get authenticated user context while traversing notifications")

			return auth.ErrNoAuthUser
		}

		// Apply the filter by modifying the query in place
		nq.Where(
			notification.Or(
				notification.UserID(caller.SubjectID),
				notification.And(
					notification.UserIDIsNil(),
					notification.OwnerIDIn(caller.OrgIDs()...),
				),
			),
		)

		return nil
	})
}
