package interceptors

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/notification"
	"github.com/theopenlane/core/v2/internal/ent/generated/predicate"
	"github.com/theopenlane/core/v2/pkg/logx"
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

		// internal operations such as delivery listeners read notifications on behalf of any user
		if caller.Has(auth.CapInternalOperation) || caller.Has(auth.CapBypassOrgFilter) {
			return nil
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
			inAppChannelPredicate(),
		)

		return nil
	})
}

// inAppChannelPredicate limits reads to notifications delivered in-app: rows with no channels set
// predate channel routing and count as in-app, otherwise the channels must include IN_APP
func inAppChannelPredicate() predicate.Notification {
	return notification.Or(
		notification.ChannelsIsNil(),
		func(s *sql.Selector) {
			s.Where(sqljson.LenEQ(notification.FieldChannels, 0))
		},
		func(s *sql.Selector) {
			s.Where(sqljson.ValueContains(notification.FieldChannels, enums.ChannelInApp.String()))
		},
	)
}
