package gala

import "context"

// DurableDispatcher dispatches envelopes to a durable transport
type DurableDispatcher interface {
	// DispatchDurable dispatches an envelope using the supplied topic policy
	DispatchDurable(context.Context, Envelope, TopicPolicy) error
}
