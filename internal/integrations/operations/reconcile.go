package operations

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
)

// ReconcileEnvelope is the durable payload for a scheduled reconciliation cycle
type ReconcileEnvelope struct {
	types.ExecutionMetadata
	// Schedule is the adaptive scheduling state carried across cycles
	Schedule gala.ScheduleState `json:"schedule"`
}

// reconcileSchemaName is the type name derived from the JSON schema reflector
var reconcileSchemaName = providerkit.SchemaID(providerkit.SchemaFrom[ReconcileEnvelope]())

var (
	// ReconcileTopic is the Gala topic name for reconciliation envelopes
	ReconcileTopic = gala.TopicName("integration." + reconcileSchemaName)
	// reconcileListenerName is the Gala listener name for reconciliation handlers
	reconcileListenerName = "integration." + reconcileSchemaName + ".handler"
)

// ReconcileHandler processes one reconciliation envelope and returns the number of
// operations dispatched (used as the delta for adaptive scheduling)
type ReconcileHandler func(context.Context, ReconcileEnvelope) (int, error)

// RegisterReconcileListener registers the Gala listener for integration reconciliation
func RegisterReconcileListener(runtime *gala.Gala, handle ReconcileHandler, schedule gala.Schedule) error {
	return RegisterScheduledListener(ScheduledListenerConfig[ReconcileEnvelope]{
		Runtime:  runtime,
		Topic:    ReconcileTopic,
		Name:     reconcileListenerName,
		Schedule: schedule,
		Handle:   handle,
		State:    func(e ReconcileEnvelope) gala.ScheduleState { return e.Schedule },
		Wrap: func(e ReconcileEnvelope, s gala.ScheduleState) ReconcileEnvelope {
			return ReconcileEnvelope{
				ExecutionMetadata: e.ExecutionMetadata,
				Schedule:          s,
			}
		},
		PrepareEmit: func(ctx context.Context, e ReconcileEnvelope) (context.Context, gala.Headers) {
			return types.WithExecutionMetadata(ctx, e.ExecutionMetadata), gala.Headers{
				Properties: e.Properties(),
			}
		},
	})
}
