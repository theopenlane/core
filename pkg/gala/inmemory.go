package gala

import (
	"context"

	"github.com/theopenlane/core/pkg/logx"
)

// dispatchFunc adapts a function to the dispatcher interface
type dispatchFunc func(context.Context, Envelope) error

// Dispatch invokes the wrapped function
func (f dispatchFunc) Dispatch(ctx context.Context, envelope Envelope) error {
	return f(ctx, envelope)
}

// NewInMemory creates a Gala runtime that dispatches envelopes asynchronously via an in-process pool.
// This is useful when listeners should not require durable River workers.
func NewInMemory() (*Gala, error) {
	return newInMemoryGala(Config{
		DispatchMode: DispatchModeInMemory,
		WorkerCount:  1,
	})
}

func newInMemoryGala(config Config) (*Gala, error) {
	g := &Gala{}
	if err := g.initialize(nil, DispatchModeInMemory); err != nil {
		return nil, err
	}

	workerCount := config.WorkerCount
	if workerCount < 1 {
		workerCount = 1
	}

	g.inMemoryPool = NewPool(
		WithWorkers(workerCount),
		WithPoolName("gala-in-memory-dispatch"),
	)

	g.dispatcher = dispatchFunc(func(ctx context.Context, envelope Envelope) error {
		pool := g.inMemoryPool
		if pool == nil {
			return g.dispatchEnvelope(ctx, envelope)
		}

		pool.Submit(func() {
			if err := g.dispatchEnvelope(ctx, envelope); err != nil {
				logx.FromContext(ctx).Warn().Err(err).Str("event_id", string(envelope.ID)).Str("topic", string(envelope.Topic)).Msg("gala in-memory listener dispatch failed")
			}
		})

		return nil
	})

	return g, nil
}
