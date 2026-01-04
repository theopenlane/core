package soiree

import "errors"

var (
	// ErrNilListener is returned when a listener is nil
	ErrNilListener = errors.New("listener cannot be nil")
	// ErrInvalidTopicName is returned when a topic name is invalid
	ErrInvalidTopicName = errors.New("invalid topic name")
	// ErrTopicNotFound is returned when a topic is not found
	ErrTopicNotFound = errors.New("topic not found")
	// ErrListenerNotFound is returned when a listener is not found
	ErrListenerNotFound = errors.New("listener not found")
	// ErrEmitterClosed is returned when the event bus is closed
	ErrEmitterClosed = errors.New("event bus is closed")
	// ErrEmitterAlreadyClosed is returned when the event bus is already closed
	ErrEmitterAlreadyClosed = errors.New("event bus is already closed")
	// errNilEventBus is returned when an event bus is nil
	errNilEventBus = errors.New("event bus is nil")
	// errMissingTypedUnwrap is returned when a typed topic is missing an unwrap helper
	errMissingTypedUnwrap = errors.New("soiree: missing unwrap helper for typed topic")
	// ErrNilPayload is returned when an event payload is nil
	ErrNilPayload = errors.New("nil payload")
	// ErrPayloadTypeMismatch is returned when an event payload type does not match the expected type
	ErrPayloadTypeMismatch = errors.New("payload type mismatch")
	// ErrEventTopicMismatch is returned when the emitted topic name disagrees with the Event.Topic() value
	ErrEventTopicMismatch = errors.New("event topic mismatch")
)
