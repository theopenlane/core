package soiree

import (
	"encoding/json"
	"fmt"
)

// TypedTopic represents a strongly typed event topic. It carries helpers that convert
// between the strongly typed payload and the internal soiree.Event representation
type TypedTopic[T any] struct {
	name          string
	wrap          func(T) Event
	unwrap        func(Event) (T, error)
	observability *ObservabilitySpec[T]
}

// ObservabilitySpec describes logging/metrics metadata for a typed topic
type ObservabilitySpec[T any] struct {
	// Operation is the operation name to record
	Operation string
	// Origin is the component emitting the observation
	Origin string
	// TriggerFunc overrides how trigger event values are derived
	TriggerFunc func(*EventContext, T) string
}

// TypedListener represents a listener that expects a strongly typed payload
type TypedListener[T any] func(*EventContext, T) error

// TypedTopicOption configures a TypedTopic
type TypedTopicOption[T any] func(*TypedTopic[T])

// WithWrap sets a custom wrap function for the typed topic
func WithWrap[T any](wrap func(T) Event) TypedTopicOption[T] {
	return func(t *TypedTopic[T]) {
		t.wrap = wrap
	}
}

// WithUnwrap sets a custom unwrap function for the typed topic
func WithUnwrap[T any](unwrap func(Event) (T, error)) TypedTopicOption[T] {
	return func(t *TypedTopic[T]) {
		t.unwrap = unwrap
	}
}

// WithObservability sets an observability spec for the typed topic
func WithObservability[T any](spec ObservabilitySpec[T]) TypedTopicOption[T] {
	return func(t *TypedTopic[T]) {
		t.observability = &spec
	}
}

// NewTypedTopic constructs a typed topic with default wrap and unwrap helpers
func NewTypedTopic[T any](name string, opts ...TypedTopicOption[T]) TypedTopic[T] {
	t := TypedTopic[T]{
		name: name,
		wrap: func(p T) Event {
			return NewBaseEvent(name, p)
		},
		unwrap: UnwrapPayload[T],
	}

	for _, opt := range opts {
		opt(&t)
	}

	return t
}

// UnwrapPayload extracts a typed payload from an event, handling JSON deserialization if needed
func UnwrapPayload[T any](event Event) (T, error) {
	var zero T

	if event == nil {
		return zero, ErrNilPayload
	}

	payload := event.Payload()
	if payload == nil {
		return zero, ErrNilPayload
	}

	typed, ok := payload.(T)
	if ok {
		return typed, nil
	}

	var raw json.RawMessage

	switch v := payload.(type) {
	case json.RawMessage:
		raw = v
	case []byte:
		raw = v
	default:
		encoded, err := json.Marshal(payload)
		if err != nil {
			return zero, fmt.Errorf("%w: expected %T, got %T", ErrPayloadTypeMismatch, zero, payload)
		}

		raw = encoded
	}

	var decoded T
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return zero, fmt.Errorf("%w: expected %T, got %T", ErrPayloadTypeMismatch, zero, payload)
	}

	event.SetPayload(decoded)

	return decoded, nil
}

// Name exposes the string representation of the topic
func (t TypedTopic[T]) Name() string {
	return t.name
}

// Wrap converts a typed payload into an Event using the topic's wrap helper.
func (t TypedTopic[T]) Wrap(payload T) (Event, error) {
	if t.wrap == nil {
		return nil, fmt.Errorf("%w: %s", errMissingTypedWrap, t.Name())
	}

	return t.wrap(payload), nil
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

			topicName := normalizeTopicName(topic.Name())
			if err := validateTopicName(topicName); err != nil {
				return "", err
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

			return bus.On(topicName, wrapped)
		},
	}
}

// Register registers the listener binding on the provided bus
func (b ListenerBinding) Register(bus *EventBus) (string, error) {
	if bus == nil {
		return "", errNilEventBus
	}
	if b.register == nil {
		return "", ErrNilListener
	}

	return b.register(bus)
}

// Observability returns the topic observability spec if configured
func (t TypedTopic[T]) Observability() (ObservabilitySpec[T], bool) {
	if t.observability == nil {
		return ObservabilitySpec[T]{}, false
	}

	return *t.observability, true
}
