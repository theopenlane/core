package soiree

import "context"

// Listener handles an event via the provided event context wrapper
type Listener func(*EventContext) error

// Hook allows callers to execute logic before or after the main listener
type Hook func(*EventContext) error

// listenerItem stores a listener along with its unique identifier and priority
type listenerItem struct {
	listener  Listener
	priority  Priority
	preHooks  []Hook
	postHooks []Hook
}

// ListenerOption configures listener behavior
type ListenerOption func(*listenerItem)

// Priority type for listener priority levels
type Priority int

const (
	Lowest Priority = iota + 1 // Lowest priority
	Low
	Normal
	High
	Highest
)

// WithPriority sets the priority of a listener
func WithPriority(priority Priority) ListenerOption {
	return func(item *listenerItem) {
		item.priority = priority
	}
}

// WithPreHooks registers hook functions that run before the listener
func WithPreHooks(hooks ...Hook) ListenerOption {
	return func(item *listenerItem) {
		item.preHooks = append(item.preHooks, hooks...)
	}
}

// WithPostHooks registers hook functions that run after the listener
func WithPostHooks(hooks ...Hook) ListenerOption {
	return func(item *listenerItem) {
		item.postHooks = append(item.postHooks, hooks...)
	}
}

// call executes the listener along with its pre- and post-hooks
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

	if event := ctx.Event(); event != nil && event.IsAborted() {
		return err
	}

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

// Context bundles the event, payload, and client for a listener
type EventContext struct {
	event     Event
	payload   any
	client    any
	hasClient bool
}

// newEventContext constructs a new EventContext from the provided event
func newEventContext(event Event) *EventContext {
	ctx := &EventContext{
		event:   event,
		payload: event.Payload(),
		client:  event.Client(),
	}

	if ctx.client != nil {
		ctx.hasClient = true
	}

	return ctx
}

// Context returns the underlying request context
func (c *EventContext) Context() context.Context {
	if c == nil || c.event == nil {
		return context.Background()
	}

	return c.event.Context()
}

// Event exposes the underlying event
func (c *EventContext) Event() Event {
	if c == nil {
		return nil
	}

	return c.event
}

// Payload returns the raw payload attached to the event
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

// Client returns the associated client if available
func (c *EventContext) Client() (any, bool) {
	if c == nil || !c.hasClient {
		return nil, false
	}

	return c.client, true
}

// Properties exposes the underlying property map, ensuring one exists
func (c *EventContext) Properties() Properties {
	if c == nil || c.event == nil {
		return NewProperties()
	}

	props := c.event.Properties()
	if props == nil {
		// Older callers sometimes left properties unset; hydrate an empty map on demand so helper
		// methods can safely mutate it and downstream listeners all see the same backing map
		props = NewProperties()
		c.event.SetProperties(props)
	}

	return props
}

// Property fetches a property by key.
func (c *EventContext) Property(key string) (any, bool) {
	props := c.Properties()
	if props == nil {
		return nil, false
	}

	val, ok := props[key]
	if !ok || val == nil {
		// Treat zero values the same way legacy code did—absence and explicit nil should both
		// short circuit so callers can distinguish “not present” from empty strings
		return nil, false
	}

	return val, true
}

// PropertyString fetches a string property by key
func (c *EventContext) PropertyString(key string) (string, bool) {
	val, ok := c.Property(key)
	if !ok {
		return "", false
	}

	str, ok := val.(string)
	if !ok {
		return "", false
	}

	return str, true
}

// SetProperty mutates a property on the event
func (c *EventContext) SetProperty(key string, value any) {
	if c == nil {
		return
	}

	c.Properties().Set(key, value)
}

// Abort marks the event as aborted
func (c *EventContext) Abort() {
	if c == nil || c.event == nil {
		return
	}

	c.event.SetAborted(true)
}

// PayloadAs attempts to cast the payload to the requested type
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

// ClientAs attempts to cast the client to the requested type
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
