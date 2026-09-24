package hooks

import (
	"context"
	"slices"

	"github.com/theopenlane/core/common/enums"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphsubscriptions"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// init registers the notification delivery listeners so gala setup picks them up automatically
func init() { registerListeners(NotificationDeliveryListeners) }

// NotificationDeliveryListeners publishes newly created in-app notifications to live subscribers;
// external channels such as email are dispatched by the notify spec that created the row
func NotificationDeliveryListeners() []gala.Registration {
	return []gala.Registration{
		entityops.MutationListener{
			Concern:    entityops.MutationConcernNotification,
			Schema:     entityops.SchemaNotification,
			Operations: []string{entityops.OpCreate},
			Caller:     internalCaller,
			Handle:     handleNotificationDelivery,
		},
	}
}

// handleNotificationDelivery publishes the notification to live subscriptions when it carries the in-app channel
func handleNotificationDelivery(inv entityops.Invocation, _ entityops.MutationPayload) error {
	notification, ok, err := entityops.LoadEntity(inv.Context, inv.EntityID, inv.Client.Notification.Get)
	if err != nil || !ok {
		return err
	}

	if !slices.Contains(notification.Channels, enums.ChannelInApp) {
		return nil
	}

	publishNotification(inv.Context, notification)

	return nil
}

// publishNotification pushes the notification to live graph subscriptions, targeting the single
// user when one is named and otherwise every session in the owning organization
func publishNotification(ctx context.Context, notification *generated.Notification) {
	manager := graphsubscriptions.GetGlobalManager()
	if manager == nil {
		return
	}

	if notification.UserID == "" && notification.OwnerID == "" {
		return
	}

	if err := manager.Publish(notification.UserID, notification.OwnerID, notification); err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("notification_id", notification.ID).Msg("notification delivery: failed publishing to subscribers")
	}
}
