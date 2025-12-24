package interceptors

import (
	"context"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/notification"
	"github.com/theopenlane/iam/auth"
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
		ac, err := auth.GetAuthenticatedUserFromContext(ctx)
		if err != nil {
			return err
		}

		// Apply the filter by modifying the query in place
		nq.Where(
			notification.Or(
				notification.UserID(ac.SubjectID),
				notification.And(
					notification.UserIDIsNil(),
					notification.OwnerIDIn(ac.OrganizationIDs...),
				),
			),
		)

		return nil
	})
}
