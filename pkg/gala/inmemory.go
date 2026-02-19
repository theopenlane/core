package gala

import "context"

// DispatchFunc adapts a function to the Dispatcher interface.
type DispatchFunc func(context.Context, Envelope) error

// Dispatch invokes the wrapped function.
func (f DispatchFunc) Dispatch(ctx context.Context, envelope Envelope) error {
	return f(ctx, envelope)
}

// NewInMemory creates a Gala runtime that dispatches envelopes immediately in-process.
// This is useful for tests that need deterministic listener execution without River workers.
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

	g.dispatcher = DispatchFunc(func(ctx context.Context, envelope Envelope) error {
		pool := g.inMemoryPool
		if pool == nil {
			return g.DispatchEnvelope(ctx, envelope)
		}

		errCh := make(chan error, 1)
		pool.Submit(func() {
			errCh <- g.DispatchEnvelope(ctx, envelope)
		})

		return <-errCh
	})

	return g, nil
}
