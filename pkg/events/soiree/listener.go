package soiree

import "context"

// Listener handles an event via the provided event context wrapper.
type Listener func(*EventContext) error

// Hook allows callers to execute logic before or after the main listener.
type Hook func(*EventContext) error

// listenerItem stores a listener along with its unique identifier and priority.
type listenerItem struct {
	listener  Listener
	priority  Priority
	preHooks  []Hook
	postHooks []Hook
}

// ListenerOption configures listener behaviour.
type ListenerOption func(*listenerItem)

// WithPriority sets the priority of a listener.
func WithPriority(priority Priority) ListenerOption {
	return func(item *listenerItem) {
		item.priority = priority
	}
}

// WithPreHooks registers hook functions that run before the listener.
// Hooks should only perform lightweight context setup or validation; they must not run business logic.
func WithPreHooks(hooks ...Hook) ListenerOption {
	return func(item *listenerItem) {
		item.preHooks = append(item.preHooks, hooks...)
	}
}

// WithPostHooks registers hook functions that run after the listener.
func WithPostHooks(hooks ...Hook) ListenerOption {
	return func(item *listenerItem) {
		item.postHooks = append(item.postHooks, hooks...)
	}
}

func (item *listenerItem) call(ctx *EventContext) error {
	for _, hook := range item.preHooks {
		if hook == nil {
			continue
		}

		if err := hook(ctx); err != nil || ctx.Event().IsAborted() {
			return err
		}
	}

	err := item.listener(ctx)

	for _, hook := range item.postHooks {
		if hook == nil {
			continue
		}

		if hookErr := hook(ctx); hookErr != nil && err == nil {
			err = hookErr
		}
	}

	return err
}

// Context bundles the event, payload, properties, and client for a listener.
type EventContext struct {
	event      Event
	payload    any
	properties PropertyView
	client     any
	hasClient  bool
}

func newEventContext(event Event) *EventContext {
	props := event.Properties()
	if props == nil {
		props = NewProperties()
		event.SetProperties(props)
	}

	ctx := &EventContext{
		event:      event,
		payload:    event.Payload(),
		properties: NewPropertyView(props),
		client:     event.Client(),
	}

	if ctx.client != nil {
		ctx.hasClient = true
	}

	return ctx
}

// Context returns the underlying request context.
func (c *EventContext) Context() context.Context {
	if c == nil || c.event == nil {
		return context.Background()
	}

	return c.event.Context()
}

// Event exposes the underlying event.
func (c *EventContext) Event() Event {
	if c == nil {
		return nil
	}

	return c.event
}

// Payload returns the raw payload attached to the event.
func (c *EventContext) Payload() any {
	if c == nil {
		return nil
	}

	return c.payload
}

func (c *EventContext) setPayload(value any) {
	if c == nil {
		return
	}

	c.payload = value
	if c.event != nil {
		c.event.SetPayload(value)
	}
}

// Client returns the associated client if available.
func (c *EventContext) Client() (any, bool) {
	if c == nil || !c.hasClient {
		return nil, false
	}

	return c.client, true
}

// Properties exposes the typed property view.
func (c *EventContext) Properties() PropertyView {
	if c == nil {
		return NewPropertyView(nil)
	}

	return c.properties
}

// SetProperty mutates a property on the event.
func (c *EventContext) SetProperty(key string, value any) {
	if c == nil {
		return
	}

	c.properties.Set(key, value)
}

// Abort marks the event as aborted.
func (c *EventContext) Abort() {
	if c == nil || c.event == nil {
		return
	}

	c.event.SetAborted(true)
}

// PayloadAs attempts to cast the payload to the requested type.
func PayloadAs[T any](ctx *EventContext) (T, bool) {
	var zero T
	if ctx == nil {
		return zero, false
	}

	value, ok := ctx.Payload().(T)
	if !ok {
		return zero, false
	}

	return value, true
}

// ClientAs attempts to cast the client to the requested type.
func ClientAs[T any](ctx *EventContext) (T, bool) {
	var zero T
	if ctx == nil {
		return zero, false
	}

	client, ok := ctx.Client()
	if !ok {
		return zero, false
	}

	typed, ok := client.(T)
	if !ok {
		return zero, false
	}

	return typed, true
}
