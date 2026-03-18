package graphapi

import (
	"context"
	"fmt"
	"time"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/notification"
	"github.com/theopenlane/core/internal/graphapi/common"
	"github.com/theopenlane/core/internal/graphsubscriptions"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/iam/auth"
)

// handleNotificationSubscription handles the subscription logic for notifications.
// It fetches existing notifications and sets up channels for real-time updates.
// This is extracted into a helper to prevent code generation from overwriting the implementation.
func (r *subscriptionResolver) handleNotificationSubscription(ctx context.Context) (<-chan *generated.Notification, error) {
	// Check if subscription manager is available
	if r.subscriptionManager == nil {
		logx.FromContext(ctx).Info().Msg("subscription manager is not initialized")
		return nil, common.ErrInternalServerError
	}

	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil {
		return nil, fmt.Errorf("failed to get user ID from context: %w", auth.ErrNoAuthUser)
	}
	userID := caller.SubjectID
	if userID == "" {
		return nil, fmt.Errorf("failed to get user ID from context: %w", auth.ErrNoAuthUser)
	}

	// Create a channel with the interface type for the subscription manager
	internalChan := make(chan graphsubscriptions.Notification, graphsubscriptions.NotificationChannelBufferSize)
	r.subscriptionManager.Subscribe(userID, internalChan)

	// Create a channel with the concrete type for the GraphQL response
	notifChan := make(chan *generated.Notification, graphsubscriptions.NotificationChannelBufferSize)

	// Fetch existing notifications for the user so they see their old notifications
	// Query filters:
	// 1. All unread notifications (read_at IS NULL)
	// 2. All read notifications created within the lookback period
	query := r.db.Notification.Query()

	// Apply filtering based on read status and lookback period
	// Use Or to combine: (unread) OR (read AND created within lookback)
	if r.notificationLookbackDays > 0 {
		// Calculate the lookback cutoff date
		lookbackCutoff := time.Now().AddDate(0, 0, -r.notificationLookbackDays)

		query = query.Where(
			notification.Or(
				// All unread notifications
				notification.ReadAtIsNil(),
				// Read notifications within the lookback period
				notification.And(
					notification.ReadAtNotNil(),
					notification.CreatedAtGTE(lookbackCutoff),
				),
			),
		)
	} else {
		// If lookback is 0 or negative, only fetch unread notifications
		query = query.Where(notification.ReadAtIsNil())
	}

	existingNotifications, err := query.All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to fetch existing notifications")
		// Continue without existing notifications - don't fail the subscription
	}

	// Forward notifications from internal channel to GraphQL channel
	go func() {
		defer close(notifChan)

		// First, send all existing notifications
		for _, existingNotif := range existingNotifications {
			select {
			case notifChan <- existingNotif:
			case <-ctx.Done():
				r.subscriptionManager.Unsubscribe(userID, internalChan)
				return
			}
		}

		// Then listen for new real-time notifications
		for {
			select {
			case <-ctx.Done():
				r.subscriptionManager.Unsubscribe(userID, internalChan)
				return
			case notif, ok := <-internalChan:
				if !ok {
					return
				}
				// Cast back to concrete type
				if concreteNotif, ok := notif.(*generated.Notification); ok {
					select {
					case notifChan <- concreteNotif:
					case <-ctx.Done():
						r.subscriptionManager.Unsubscribe(userID, internalChan)
						return
					}
				}
			}
		}
	}()

	return notifChan, nil
}
