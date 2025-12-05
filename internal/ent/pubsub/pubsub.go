package pubsub

import "github.com/theopenlane/core/internal/ent/generated"

// NotificationPublisher is an interface for publishing notifications to subscribers
type NotificationPublisher interface {
	Publish(userID string, notification *generated.Notification) error
}
