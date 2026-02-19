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
	return newInMemoryGala()
}

func newInMemoryGala() (*Gala, error) {
	g := &Gala{}
	if err := g.initialize(nil, DispatchModeInMemory); err != nil {
		return nil, err
	}

	g.dispatcher = DispatchFunc(func(ctx context.Context, envelope Envelope) error {
		return g.DispatchEnvelope(ctx, envelope)
	})

	return g, nil
}
