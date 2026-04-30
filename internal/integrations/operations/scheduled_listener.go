package operations

import (
	"context"

	"github.com/riverqueue/river"

	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// ScheduledListenerConfig defines the registration parameters for a
// self-sustaining Gala listener with adaptive scheduling
type ScheduledListenerConfig[T any] struct {
	// Runtime is the Gala instance to register on
	Runtime *gala.Gala
	// Topic is the Gala topic name
	Topic gala.TopicName
	// Name is the stable listener name
	Name string
	// Schedule controls adaptive interval computation
	Schedule gala.Schedule
	// Handle is the handler invoked each cycle, returning the delta for scheduling
	Handle func(context.Context, T) (int, error)
	// State extracts the ScheduleState from the envelope
	State func(T) gala.ScheduleState
	// Wrap builds a new envelope carrying the updated ScheduleState
	Wrap func(T, gala.ScheduleState) T
	// PrepareEmit optionally enriches the context and headers before re-emitting
	PrepareEmit func(context.Context, T) (context.Context, gala.Headers)
}

// RegisterScheduledListener registers a self-sustaining Gala listener that
// processes one cycle, computes the next adaptive interval, and re-emits
func RegisterScheduledListener[T any](cfg ScheduledListenerConfig[T]) error {
	if cfg.Runtime == nil {
		return ErrGalaRequired
	}

	_, err := gala.RegisterListeners(cfg.Runtime.Registry(), gala.Definition[T]{
		Topic: gala.Topic[T]{Name: cfg.Topic},
		Name:  cfg.Name,
		Handle: func(ctx gala.HandlerContext, envelope T) error {
			delta, execErr := cfg.Handle(ctx.Context, envelope)

			if execErr != nil {
				state := cfg.State(envelope)
				logx.FromContext(ctx.Context).Warn().Err(execErr).Int("error_streak", state.ErrorStreak+1).Msg("scheduled listener cycle failed, scheduling retry with backoff")
			}

			next := cfg.Schedule.Next(cfg.State(envelope), delta, execErr)
			scheduledAt := next.NextScheduledAt()

			emitCtx := ctx.Context
			headers := gala.Headers{}
			if cfg.PrepareEmit != nil {
				emitCtx, headers = cfg.PrepareEmit(ctx.Context, envelope)
			}

			headers.ScheduledAt = &scheduledAt

			receipt := cfg.Runtime.EmitWithHeaders(emitCtx, cfg.Topic, cfg.Wrap(envelope, next), headers)

			if execErr != nil {
				return river.JobCancel(execErr)
			}

			return receipt.Err
		},
	})

	return err
}
