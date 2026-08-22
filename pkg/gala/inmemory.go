package gala

import (
	"context"

	"github.com/theopenlane/core/pkg/logx"
)

func newInMemoryGala(config Config) (*Gala, error) {
	g := &Gala{}
	if err := g.initialize(DispatchModeInMemory); err != nil {
		return nil, err
	}

	g.inMemoryPool = NewPool(
		WithWorkers(config.WorkerCount),
		WithPoolName("gala-in-memory-dispatch"),
	)

	return g, nil
}

// dispatchInMemory submits the envelope to the in-memory pool, dispatching inline when no
// pool is configured
func (g *Gala) dispatchInMemory(ctx context.Context, envelope Envelope) error {
	if g.inMemoryPool == nil {
		return g.dispatchEnvelope(ctx, envelope)
	}

	g.inMemoryPool.Submit(func() {
		if err := g.dispatchEnvelope(ctx, envelope); err != nil {
			logx.FromContext(ctx).Warn().Err(err).Str("event_id", string(envelope.ID)).Str("topic", string(envelope.Topic)).Msg("gala in-memory listener dispatch failed")
		}
	})

	return nil
}
