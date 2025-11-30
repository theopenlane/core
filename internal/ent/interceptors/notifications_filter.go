// core/internal/ent/interceptors/notification_filter.go

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
		subjectID, err := auth.GetSubjectIDFromContext(ctx)
		if err != nil {
			// If no auth context, let it proceed (might be internal query)
			return nil
		}

		orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
		if err != nil {
			orgIDs = []string{} // Default to empty if org IDs not found
		}

		// Apply the filter by modifying the query in place
		nq.Where(
			notification.Or(
				notification.UserID(subjectID),
				notification.And(
					notification.UserIDIsNil(),
					notification.OwnerIDIn(orgIDs...),
				),
			),
		)

		return nil
	})
}
