package soiree

import (
	"context"
)

// eventStore defines methods for persisting events and listener results
type eventStore interface {
	// SaveEvent persists the incoming event before processing
	SaveEvent(Event) error
	// SaveHandlerResult persists the result of a listener processing an event
	SaveHandlerResult(Event, string, error) error
}

// eventQueue extends eventStore with the ability to dequeue events for processing
type eventQueue interface {
	eventStore
	// DequeueEvent retrieves the next event from the queue for processing
	DequeueEvent(context.Context) (Event, error)
}

// handlerResultDeduper optionally enables idempotent handler execution across event replays
type handlerResultDeduper interface {
	HandlerSucceeded(context.Context, string, string) (bool, error)
}

// StoredResult holds the outcome of a listener processing an event
type StoredResult struct {
	// Topic is the topic of the event that was processed
	Topic string `json:"topic"`
	// EventID is the unique idempotency key for the event, when available
	EventID string `json:"event_id,omitempty"`
	// HandlerID is the unique identifier of the listener that processed the event
	HandlerID string `json:"handler_id"`
	// Error is the error encountered while processing the event
	Error string `json:"error,omitempty"`
}
