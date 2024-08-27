package events

import (
	"context"
)

// This file currently contains unused interfaces and types that will be called / integrated in the future

// EventConsumer is the interface for consuming events
type EventConsumer interface {
	Subscribe(ctx context.Context) error
	Close(context.Context) error
}

// EventPublisher is the interface for publishing events
type EventPublisher interface {
	StartPublisher(context.Context) error
	Publish(ctx context.Context, topic string, payload interface{}) error
	Close(context.Context) error
}

// Properties is a map of properties to set on an event as a wrapper
type Properties map[string]interface{}

// NewProperties creates a new Properties map
func NewProperties() Properties {
	return make(Properties, 10) //nolint:mnd
}

// Set sets a property on the Properties map
func (p Properties) Set(name string, value interface{}) Properties {
	p[name] = value

	return p
}
