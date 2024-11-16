package soiree

import "errors"

var (
	// ErrNilListener is returned when a listener is nil
	ErrNilListener = errors.New("listener cannot be nil")
	// ErrInvalidTopicName is returned when a topic name is invalid
	ErrInvalidTopicName = errors.New("invalid topic name")
	// ErrInvalidPriority is returned when a priority is invalid
	ErrInvalidPriority = errors.New("invalid priority")
	// ErrTopicNotFound is returned when a listener option is invalid
	ErrTopicNotFound = errors.New("topic not found")
	// ErrListenerNotFound is returned when a listener is not found
	ErrListenerNotFound = errors.New("listener not found")
	// ErrEventProcessingAborted is returned when event processing is aborted
	ErrEventProcessingAborted = errors.New("event processing aborted")
	// ErrEmitterClosed is returned when the soiree is closed
	ErrEmitterClosed = errors.New("soiree is closed")
	// ErrEmitterAlreadyClosed is returned when the soiree is already closed
	ErrEmitterAlreadyClosed = errors.New("soiree is already closed")
	// ErrPayloadisNotEvent is returned when the payload is not an event
	ErrPayloadisNotEvent = errors.New("payload is not an event")
)
