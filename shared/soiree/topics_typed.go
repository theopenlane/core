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
	register func(*EventPool) (string, error)
}

// BindListener produces a binding that can be registered on an EventPool in a batch
func BindListener[T any](topic TypedTopic[T], listener TypedListener[T], opts ...ListenerOption) ListenerBinding {
	return ListenerBinding{
		register: func(pool *EventPool) (string, error) {
			if pool == nil {
				return "", errNilEventPool
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
				payload, err := topic.unwrap(ctx.Event())
				if err != nil {
					return err
				}

				ctx.setPayload(payload)

				return listener(ctx, payload)
			}

			return pool.On(topic.Name(), wrapped, opts...)
		},
	}
}

// registerWith registers the listener binding on the provided pool
func (b ListenerBinding) registerWith(pool *EventPool) (string, error) {
	if pool == nil {
		return "", errNilEventPool
	}
	if b.register == nil {
		return "", ErrNilListener
	}

	return b.register(pool)
}

// Register registers the listener binding on the provided pool
func (b ListenerBinding) Register(pool *EventPool) (string, error) {
	return b.registerWith(pool)
}
