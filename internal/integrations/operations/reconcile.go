package operations

import (
	"context"
	"errors"

	"github.com/riverqueue/river"

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
	if runtime == nil {
		return ErrGalaRequired
	}

	topic := gala.Topic[ReconcileEnvelope]{Name: ReconcileTopic}

	_, err := gala.RegisterListeners(runtime.Registry(), gala.Definition[ReconcileEnvelope]{
		Topic: topic,
		Name:  reconcileListenerName,
		Handle: func(ctx gala.HandlerContext, envelope ReconcileEnvelope) error {
			delta, execErr := handle(ctx.Context, envelope)

			if execErr != nil {
				if ent.IsNotFound(execErr) {
					logx.FromContext(ctx.Context).Error().Err(execErr).Msg("integration not found, not queuing")
					return river.JobCancel(execErr)
				}

				if errors.Is(execErr, ErrOperationDisabled) {
					logx.FromContext(ctx.Context).Info().Str("integration_id", envelope.IntegrationID).Str("operation", envelope.Operation).Msg("operation disabled, stopping reconcile cycle")
					return river.JobCancel(execErr)
				}

				logx.FromContext(ctx.Context).Warn().Err(execErr).Str("integration_id", envelope.IntegrationID).Str("operation", envelope.Operation).Int("error_streak", envelope.Schedule.ErrorStreak+1).Msg("reconcile cycle failed, scheduling retry with backoff")
			}

			effectiveSchedule := schedule
			if reg != nil {
				if op, err := reg.Operation(envelope.DefinitionID, envelope.Operation); err == nil && op.ReconcileSchedule != nil {
					effectiveSchedule = *op.ReconcileSchedule
				}
			}

			next := effectiveSchedule.Next(envelope.Schedule, delta, execErr)
			scheduledAt := next.NextScheduledAt()
			emitCtx := types.WithExecutionMetadata(ctx.Context, envelope.ExecutionMetadata)
			receipt := runtime.EmitWithHeaders(emitCtx, ReconcileTopic, ReconcileEnvelope{
				ExecutionMetadata: envelope.ExecutionMetadata,
				Schedule:          next,
			}, gala.Headers{
				ScheduledAt: &scheduledAt,
				Properties:  envelope.Properties(),
				Tags:        types.GetTagsForExecutionMetadata(envelope.ExecutionMetadata),
			})

			if execErr != nil {
				return river.JobCancel(execErr)
			}

			return receipt.Err
		},
	})

	return err
}
