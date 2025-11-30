package interceptors

import (
	"context"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/notification"
	"github.com/theopenlane/iam/auth"
)

// NotificationQueryFilter automatically filters notifications based on user context
func NotificationQueryFilter() generated.Interceptor {
	return generated.InterceptFunc(func(next generated.Querier) generated.Querier {
		return generated.QuerierFunc(func(ctx context.Context, q generated.Query) (generated.Value, error) {
			// Only apply to Notification queries
			nq, ok := q.(*generated.NotificationQuery)
			if !ok {
				return next.Query(ctx, q)
			}

			// Get user info from context
			subjectID, err := auth.GetSubjectIDFromContext(ctx)
			if err != nil {
				// If no auth context, let it proceed (might be internal query)
				return next.Query(ctx, q)
			}

			orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
			if err != nil {
				orgIDs = []string{} // Default to empty if org IDs not found
			}

			// Apply the filter
			nq = nq.Where(
				notification.Or(
					notification.UserID(subjectID),
					notification.And(
						notification.UserIDIsNil(),
						notification.OwnerIDIn(orgIDs...),
					),
				),
			)

			return next.Query(ctx, nq)
		})
	})
}
