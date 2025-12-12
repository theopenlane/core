package graphapi

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/notification"
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
		return nil, ErrInternalServerError
	}

	userID, err := auth.GetSubjectIDFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user ID from context: %w", err)
	}

	// Create a channel with the interface type for the subscription manager
	internalChan := make(chan graphsubscriptions.Notification, graphsubscriptions.TaskChannelBufferSize)
	r.subscriptionManager.Subscribe(userID, internalChan)

	// Create a channel with the concrete type for the GraphQL response
	notifChan := make(chan *generated.Notification, graphsubscriptions.TaskChannelBufferSize)

	// Fetch existing notifications for the user so they see their old notifications
	existingNotifications, err := r.db.Notification.Query().
		Where(notification.UserIDEQ(userID)).
		All(ctx)
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
