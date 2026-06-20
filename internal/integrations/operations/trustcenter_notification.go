package operations

import (
	"context"

	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
)

// TrustCenterNotificationEnvelope is the durable payload for a trust center notification polling cycle
type TrustCenterNotificationEnvelope struct {
	// Schedule is the adaptive scheduling state carried across polling cycles
	Schedule gala.ScheduleState `json:"schedule"`
}

// trustCenterNotificationSchemaName is the type name derived from the JSON schema reflector
var trustCenterNotificationSchemaName = jsonx.SchemaID(jsonx.SchemaFrom[TrustCenterNotificationEnvelope]())

var (
	// TrustCenterNotificationTopic is the Gala topic name for trust center notification polling
	TrustCenterNotificationTopic = gala.TopicName("trustcenter.notification." + trustCenterNotificationSchemaName)
	// trustCenterNotificationListenerName is the Gala listener name for the trust center notification handler
	trustCenterNotificationListenerName = "trustcenter.notification." + trustCenterNotificationSchemaName + ".handler"
)

// TrustCenterNotificationHandler processes one polling cycle and returns the number of notifications
// dispatched (used as the delta for adaptive scheduling)
type TrustCenterNotificationHandler func(context.Context, TrustCenterNotificationEnvelope) (int, error)

// RegisterTrustCenterNotificationListener registers the Gala listener for trust center notification polling
func RegisterTrustCenterNotificationListener(runtime *gala.Gala, handle TrustCenterNotificationHandler, schedule gala.Schedule) error {
	return RegisterScheduledListener(ScheduledListenerConfig[TrustCenterNotificationEnvelope]{
		Runtime:  runtime,
		Topic:    TrustCenterNotificationTopic,
		Name:     trustCenterNotificationListenerName,
		Schedule: schedule,
		Handle:   handle,
		State:    func(e TrustCenterNotificationEnvelope) gala.ScheduleState { return e.Schedule },
		Wrap: func(_ TrustCenterNotificationEnvelope, s gala.ScheduleState) TrustCenterNotificationEnvelope {
			return TrustCenterNotificationEnvelope{Schedule: s}
		},
	})
}
