package hooks

import (
	"github.com/theopenlane/core/internal/ent/notifications"
	"github.com/theopenlane/core/pkg/gala"
)

// RegisterGalaNotificationListeners registers notification mutation listeners on Gala.
func RegisterGalaNotificationListeners(registry *gala.Registry) ([]gala.ListenerID, error) {
	return notifications.RegisterGalaListeners(registry)
}
