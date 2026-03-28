package operations

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
)

// ReconcileEnvelope is the durable payload for a scheduled reconciliation cycle
type ReconcileEnvelope struct {
	// InstallationID is the target installation identifier
	InstallationID string `json:"installationId"`
	// DefinitionID is the integration definition identifier
	DefinitionID string `json:"definitionId"`
	// Schedule is the adaptive scheduling state carried across cycles
	Schedule gala.ScheduleState `json:"schedule"`
}

var (
	// reconcileTopic is the Gala topic name for reconciliation envelopes
	reconcileTopic = types.TopicFromType[ReconcileEnvelope]()
	// reconcileListenerName is the Gala listener name for reconciliation handlers
	reconcileListenerName = types.ListenerFromType[ReconcileEnvelope]()
)

// ReconcileHandler processes one reconciliation envelope and returns the number of
// operations dispatched (used as the delta for adaptive scheduling)
type ReconcileHandler func(context.Context, ReconcileEnvelope) (int, error)

// RegisterReconcileListener registers the Gala listener for integration reconciliation
func RegisterReconcileListener(runtime *gala.Gala, handle ReconcileHandler, schedule gala.Schedule) error {
	if runtime == nil {
		return ErrGalaRequired
	}

	topic := gala.Topic[ReconcileEnvelope]{Name: reconcileTopic}

	_, err := gala.RegisterListeners(runtime.Registry(), gala.Definition[ReconcileEnvelope]{
		Topic: topic,
		Name:  reconcileListenerName,
		Handle: func(ctx gala.HandlerContext, envelope ReconcileEnvelope) error {
			delta, execErr := handle(ctx.Context, envelope)

			next := schedule.Next(envelope.Schedule, delta, execErr)
			scheduledAt := next.NextScheduledAt()

			receipt := runtime.EmitWithHeaders(ctx.Context, reconcileTopic, ReconcileEnvelope{
				InstallationID: envelope.InstallationID,
				DefinitionID:   envelope.DefinitionID,
				Schedule:       next},
				gala.Headers{ScheduledAt: &scheduledAt,
					Properties: map[string]string{
						"installation_id": envelope.InstallationID,
						"definition_id":   envelope.DefinitionID,
					},
				})

			return receipt.Err
		},
	})

	return err
}
