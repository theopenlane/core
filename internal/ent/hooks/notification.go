package hooks

import (
	"context"
	"fmt"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/graphsubscriptions"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/logx"
)

// HookNotification runs on notification mutations to validate channels
func HookNotification() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return hook.NotificationFunc(func(ctx context.Context, m *generated.NotificationMutation) (generated.Value, error) {
			// Validate channels using m.Channels()
			if channels, ok := m.Channels(); ok {
				if err := isValidChannels(channels); err != nil {
					return nil, err
				}
			}

			// Validate appended channels using m.AppendedChannels()
			if appendedChannels, ok := m.AppendedChannels(); ok {
				if err := isValidChannels(appendedChannels); err != nil {
					return nil, err
				}
			}

			return next.Mutate(ctx, m)
		})
	},
		hook.And(
			hook.Or(
				hook.HasFields("channels"),
				hook.HasAddedFields("channels"),
			),
			hook.HasOp(ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate),
		),
	)
}

// isValidChannels validates that all channels in the slice are valid
func isValidChannels(channels []enums.Channel) error {
	if len(channels) == 0 {
		return nil
	}

	validChannels := enums.Channel("").Values()
	validMap := make(map[string]bool)
	for _, v := range validChannels {
		validMap[v] = true
	}

	for _, ch := range channels {
		if !validMap[string(ch)] {
			return fmt.Errorf("%w: %s", ErrInvalidChannel, ch)
		}
	}

	return nil
}

// HookNotificationPublish runs after notification creation to publish to subscribers
func HookNotificationPublish() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.NotificationFunc(func(ctx context.Context, m *generated.NotificationMutation) (generated.Value, error) {
			// Execute the mutation first
			val, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			// After successful creation, publish to subscription manager
			notification, ok := val.(*generated.Notification)
			if !ok {
				logx.FromContext(ctx).Warn().Msg("notification hook: value is not a notification")
				return val, nil
			}

			logx.FromContext(ctx).Debug().Str("notification_id", notification.ID).Msg("notification hook: notification created, attempting to publish")

			// Get the global subscription manager
			manager := graphsubscriptions.GetGlobalManager()
			if manager == nil {
				// No subscription manager configured, skip publishing
				logx.FromContext(ctx).Debug().Msg("notification hook: subscription manager is nil, skipping publish")
				return val, nil
			}

			// Get the user ID to publish to
			userID := notification.UserID
			if userID == "" {
				// No specific user, skip publishing
				logx.FromContext(ctx).Debug().Str("notification_id", notification.ID).Msg("notification hook: userID is empty, skipping publish")
				return val, nil
			}

			logx.FromContext(ctx).Debug().Str("user_id", userID).Str("notification_id", notification.ID).Msg("notification hook: publishing to subscription manager")

			// Publish the notification to subscribers
			if err := manager.Publish(userID, notification); err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("user_id", userID).Msg("failed to publish notification to subscribers")
			}

			return val, nil
		})
	}, ent.OpCreate)
}
