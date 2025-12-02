package soiree

import (
	"context"
	"sync"
)

// EventStore defines methods for persisting events and listener results
type EventStore interface {
	// SaveEvent persists the incoming event before processing
	SaveEvent(Event) error
	// SaveHandlerResult persists the result of a listener processing an event
	SaveHandlerResult(Event, string, error) error
}

// EventQueue extends EventStore with the ability to dequeue events for processing
type EventQueue interface {
	// EventStore embeds EventStore to ensure all EventQueue implementations also implement EventStore
	EventStore
	// DequeueEvent retrieves the next event from the queue for processing
	DequeueEvent(context.Context) (Event, error)
}

// StoredResult holds the outcome of a listener processing an event
type StoredResult struct {
	// Topic is the topic of the event that was processed
	Topic string
	// HandlerID is the unique identifier of the listener that processed the event
	HandlerID string
	// Error is the error encountered while processing the event
	Error string
}

// InMemoryStore is an in-memory implementation of EventStore
type InMemoryStore struct {
	// mu is a mutex to ensure thread-safe access to the store
	mu sync.Mutex
	// events holds the events that have been saved
	events []Event
	// results holds the results of event processing by listeners
	results []StoredResult
}

// NewInMemoryStore creates a new InMemoryStore
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{}
}

// SaveEvent stores the event in memory
func (s *InMemoryStore) SaveEvent(e Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.events = append(s.events, e)

	return nil
}

// SaveHandlerResult stores the result of a listener processing an event
func (s *InMemoryStore) SaveHandlerResult(e Event, handlerID string, err error) error {
	s.mu.Lock()

	defer s.mu.Unlock()

	result := StoredResult{Topic: e.Topic(), HandlerID: handlerID}
	if err != nil {
		result.Error = err.Error()
	}

	s.results = append(s.results, result)

	return nil
}

// Events returns the stored events
func (s *InMemoryStore) Events() []Event {
	s.mu.Lock()
	defer s.mu.Unlock()

	return append([]Event(nil), s.events...)
}

// Results returns the stored results
func (s *InMemoryStore) Results() []StoredResult {
	s.mu.Lock()
	defer s.mu.Unlock()

	return append([]StoredResult(nil), s.results...)
}
