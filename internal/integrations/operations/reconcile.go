package operations

import (
	"context"
	"errors"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
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
func RegisterReconcileListener(runtime *gala.Gala, reg *registry.Registry, handle ReconcileHandler, schedule gala.Schedule) error {
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
				Tags:       types.GetTagsForExecutionMetadata(e.ExecutionMetadata),
			}
		},
		ShouldCancel: func(ctx context.Context, e ReconcileEnvelope, err error) bool {
			if ent.IsNotFound(err) {
				logx.FromContext(ctx).Error().Err(err).Msg("integration not found, not queuing")
				return true
			}

			if errors.Is(err, ErrOperationDisabled) {
				logx.FromContext(ctx).Info().Str("integration_id", e.IntegrationID).Str("operation", e.Operation).Msg("operation disabled, stopping reconcile cycle")
				return true
			}

			return false
		},
		ScheduleOverride: func(e ReconcileEnvelope) *gala.Schedule {
			if reg == nil {
				return nil
			}

			op, err := reg.Operation(e.DefinitionID, e.Operation)
			if err != nil {
				return nil
			}

			return op.ReconcileSchedule
		},
	})
}
