package operations

import (
	"context"
	"errors"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// ReconcileEnvelope is the durable payload for one recurring operation cycle, either
// installation-bound (IntegrationID set) or runtime-bound (Runtime true); the type name
// is the durable topic identity and must not change
type ReconcileEnvelope struct {
	types.ExecutionMetadata
	// Schedule is the adaptive scheduling state carried across cycles
	Schedule gala.ScheduleState `json:"schedule"`
}

// reconcileSchemaName is the type name derived from the JSON schema reflector
var reconcileSchemaName = jsonx.SchemaID(jsonx.SchemaFrom[ReconcileEnvelope]())

var (
	// ReconcileTopic is the Gala topic name for reconciliation envelopes
	ReconcileTopic = gala.TopicName("integration." + reconcileSchemaName)
	// reconcileListenerName is the Gala listener name for reconciliation handlers
	reconcileListenerName = "integration." + reconcileSchemaName + ".handler"
)

// ReconcileHandler processes one recurring cycle envelope and returns the cycle delta
// (used for adaptive scheduling)
type ReconcileHandler func(context.Context, ReconcileEnvelope) (int, error)

// RegisterReconcileListener registers the Gala listener driving every recurring operation
// cycle: installation-bound reconciliation and runtime-bound scheduled operations
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
			// not-found is terminal only for installation-bound cycles; runtime sweeps
			// surface joined per-item errors that may wrap not-found
			if e.IntegrationID != "" && ent.IsNotFound(err) {
				logx.FromContext(ctx).Error().Err(err).Msg("integration not found, not queuing")
				return true
			}

			if errors.Is(err, registry.ErrDefinitionNotFound) || errors.Is(err, registry.ErrOperationNotFound) {
				logx.FromContext(ctx).Error().Err(err).Str("definition_id", e.DefinitionID).Str("operation", e.Operation).Msg("operation no longer registered, stopping cycle")
				return true
			}

			if errors.Is(err, ErrOperationDisabled) {
				logx.FromContext(ctx).Info().Str("integration_id", e.IntegrationID).Str("operation", e.Operation).Msg("operation disabled, stopping cycle")
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

			return op.Schedule
		},
	})
}
