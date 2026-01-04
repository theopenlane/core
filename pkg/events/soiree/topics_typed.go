package soiree

import "fmt"

// TypedTopic represents a strongly typed event topic. It carries helpers that convert
// between the strongly typed payload and the internal soiree.Event representation
type TypedTopic[T any] struct {
	name   string
	wrap   func(T) Event
	unwrap func(Event) (T, error)
}

// TypedListener represents a listener that expects a strongly typed payload
type TypedListener[T any] func(*EventContext, T) error

// NewTypedTopic constructs a new typed topic with custom conversion helpers
func NewTypedTopic[T any](name string, wrap func(T) Event, unwrap func(Event) (T, error)) TypedTopic[T] {
	return TypedTopic[T]{
		name:   name,
		wrap:   wrap,
		unwrap: unwrap,
	}
}

// Name exposes the string representation of the topic
func (t TypedTopic[T]) Name() string {
	return t.name
}

// ListenerBinding encapsulates the registration of a listener against a topic
type ListenerBinding struct {
	register func(*EventBus) (string, error)
}

// BindListener produces a binding that can be registered on an EventBus
func BindListener[T any](topic TypedTopic[T], listener TypedListener[T]) ListenerBinding {
	return ListenerBinding{
		register: func(bus *EventBus) (string, error) {
			if bus == nil {
				return "", errNilEventBus
			}

			if listener == nil {
				return "", ErrNilListener
			}

			if !isValidTopicName(topic.Name()) {
				return "", ErrInvalidTopicName
			}

			if topic.unwrap == nil {
				return "", fmt.Errorf("%w: %s", errMissingTypedUnwrap, topic.Name())
			}

			wrapped := func(ctx *EventContext) error {
				payload, err := topic.unwrap(ctx.event)
				if err != nil {
					return err
				}
				return listener(ctx, payload)
			}

			return bus.On(topic.Name(), wrapped)
		},
	}
}

// registerWith registers the listener binding on the provided bus
func (b ListenerBinding) registerWith(bus *EventBus) (string, error) {
	if bus == nil {
		return "", errNilEventBus
	}
	if b.register == nil {
		return "", ErrNilListener
	}

	return b.register(bus)
}

// Register registers the listener binding on the provided bus
func (b ListenerBinding) Register(bus *EventBus) (string, error) {
	return b.registerWith(bus)
}
