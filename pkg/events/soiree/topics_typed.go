package soiree

// TypedTopic represents a strongly typed event topic. It carries helpers that convert
// between the strongly typed payload and the internal soiree.Event representation.
type TypedTopic[T any] struct {
	name   string
	wrap   func(T) Event
	unwrap func(Event) (T, error)
}

// TypedListener represents a listener that expects a strongly typed payload.
type TypedListener[T any] func(*EventContext, T) error

// NewTypedTopic constructs a new typed topic with custom conversion helpers.
func NewTypedTopic[T any](name string, wrap func(T) Event, unwrap func(Event) (T, error)) TypedTopic[T] {
	return TypedTopic[T]{
		name:   name,
		wrap:   wrap,
		unwrap: unwrap,
	}
}

// NewEventTopic creates a typed topic that uses the raw soiree.Event interface as the payload type.
func NewEventTopic(name string) TypedTopic[Event] {
	return NewTypedTopic(
		name,
		func(e Event) Event { return e },
		func(e Event) (Event, error) { return e, nil },
	)
}

// Name exposes the string representation of the topic.
func (t TypedTopic[T]) Name() string {
	return t.name
}

// ListenerBinding encapsulates the registration of a listener against a topic.
type ListenerBinding struct {
	register func(*EventPool) (string, error)
}

// BindListener produces a binding that can be registered on an EventPool in a batch.
func BindListener[T any](topic TypedTopic[T], listener TypedListener[T], opts ...ListenerOption) ListenerBinding {
	return ListenerBinding{
		register: func(pool *EventPool) (string, error) {
			return OnTopic(pool, topic, listener, opts...)
		},
	}
}

func (b ListenerBinding) registerWith(pool *EventPool) (string, error) {
	if pool == nil {
		return "", errNilEventPool
	}
	if b.register == nil {
		return "", ErrNilListener
	}

	return b.register(pool)
}
