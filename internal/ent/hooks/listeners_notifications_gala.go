package hooks

import (
	"github.com/theopenlane/core/internal/ent/notifications"
	"github.com/theopenlane/core/pkg/gala"
)

// RegisterGalaNotificationListeners registers notification mutation listeners on Gala.
func RegisterGalaNotificationListeners(g *gala.Gala) ([]gala.ListenerID, error) {
	return notifications.RegisterGalaListeners(g)
}
