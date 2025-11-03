package soiree

import "context"

// Listener is a function type that can handle events of any type
// Listener takes an `Event` as a parameter and returns an `error`. This
// allows you to define functions that conform to this specific signature, making it easier to work
// with event listeners in the other parts of the code
type Listener func(Event) error

// listenerItem stores a listener along with its unique identifier and priority
type listenerItem struct {
	listener Listener
	priority Priority
	client   any
	mws      []ListenerMiddleware
}

// ListenerOption is a function type that configures listener behavior
type ListenerOption func(*listenerItem)

// WithPriority sets the priority of a listener
func WithPriority(priority Priority) ListenerOption {
	return func(item *listenerItem) {
		item.priority = priority
	}
}

// WithClient sets the client of a listener
func WithListenerClient(client any) ListenerOption {
	return func(item *listenerItem) {
		item.client = client
	}
}

// ListenerMiddleware allows decorating listener execution (e.g. retries, metrics).
type ListenerMiddleware func(Listener) Listener

// WithMiddleware attaches one or more middleware to the listener.
func WithMiddleware(middlewares ...ListenerMiddleware) ListenerOption {
	return func(item *listenerItem) {
		item.mws = append(item.mws, middlewares...)
	}
}

func applyListenerMiddlewares(base Listener, middlewares []ListenerMiddleware) Listener {
	wrapped := base
	for i := len(middlewares) - 1; i >= 0; i-- {
		if mw := middlewares[i]; mw != nil {
			wrapped = mw(wrapped)
		}
	}

	return wrapped
}

// ContextListener handles events using a typed client and payload wrapper.
type ContextListener[C any, T any] func(*ListenerContext[C, T]) error

// ListenerContext bundles the event, payload, properties, and typed client for a listener.
type ListenerContext[C any, T any] struct {
	event      Event
	payload    T
	properties PropertyView
	client     C
	hasClient  bool
}

func newListenerContext[C any, T any](event Event, payload T) *ListenerContext[C, T] {
	props := event.Properties()
	if props == nil {
		props = NewProperties()
		event.SetProperties(props)
	}

	ctx := &ListenerContext[C, T]{
		event:      event,
		payload:    payload,
		properties: NewPropertyView(props),
	}

	if raw := event.Client(); raw != nil {
		if typed, ok := raw.(C); ok {
			ctx.client = typed
			ctx.hasClient = true
		}
	}

	return ctx
}

// Context exposes the underlying event context.
func (c *ListenerContext[C, T]) Context() context.Context {
	if c == nil || c.event == nil {
		return context.Background()
	}

	return c.event.Context()
}

// Event returns the underlying event.
func (c *ListenerContext[C, T]) Event() Event {
	if c == nil {
		return nil
	}

	return c.event
}

// Payload returns the strongly typed payload.
func (c *ListenerContext[C, T]) Payload() T {
	var zero T
	if c == nil {
		return zero
	}

	return c.payload
}

// Client returns the strongly typed client if available.
func (c *ListenerContext[C, T]) Client() (C, bool) {
	var zero C
	if c == nil || !c.hasClient {
		return zero, false
	}

	return c.client, true
}

// MustClient returns the strongly typed client or panics if unavailable.
func (c *ListenerContext[C, T]) MustClient() C {
	client, ok := c.Client()
	if !ok {
		panic("soiree: listener client is unavailable")
	}

	return client
}

// Properties exposes the typed property view.
func (c *ListenerContext[C, T]) Properties() PropertyView {
	if c == nil {
		return NewPropertyView(nil)
	}

	return c.properties
}

// SetProperty mutates a property on the event.
func (c *ListenerContext[C, T]) SetProperty(key string, value any) {
	if c == nil {
		return
	}

	c.properties.Set(key, value)
}

// Abort marks the event as aborted.
func (c *ListenerContext[C, T]) Abort() {
	if c == nil || c.event == nil {
		return
	}

	c.event.SetAborted(true)
}
