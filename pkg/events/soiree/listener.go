package soiree

import (
	"context"
)

// Listener handles an event via the provided event context wrapper
type Listener func(*EventContext) error

// listenerItem stores a listener along with its unique identifier
type listenerItem struct {
	listener Listener
}

func (item *listenerItem) call(ctx *EventContext) error {
	return item.listener(ctx)
}

// EventContext bundles the event, payload, and client for a listener
type EventContext struct {
	event     Event
	client    any
	hasClient bool
}

func newEventContext(event Event) *EventContext {
	ctx := &EventContext{
		event:  event,
		client: event.Client(),
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

// Properties exposes the underlying property map
func (c *EventContext) Properties() Properties {
	if c == nil || c.event == nil {
		return NewProperties()
	}

	props := c.event.Properties()
	if props == nil {
		props = NewProperties()
		c.event.SetProperties(props)
	}
	return props
}

// Property fetches a property by key
func (c *EventContext) Property(key string) (any, bool) {
	props := c.Properties()
	if props == nil {
		return nil, false
	}

	val, ok := props[key]
	if !ok || val == nil {
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

// ClientAs attempts to cast the client to the requested type
func ClientAs[T any](ctx *EventContext) (T, bool) {
	var zero T
	if ctx == nil || !ctx.hasClient {
		return zero, false
	}

	typed, ok := ctx.client.(T)
	if !ok {
		return zero, false
	}

	return typed, true
}
