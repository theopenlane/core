package soiree

import "sync"

// Event is an interface representing the structure of an instance of an event
type Event interface {
	// Topic returns the event's topic
	Topic() string
	// Payload returns the event's payload
	Payload() interface{}
	// Properties returns the event's properties
	Properties() Properties
	// SetPayload sets the event's payload
	SetPayload(interface{})
	// SetProperties sets the event's properties
	SetProperties(Properties)
	// SetAborted sets the event's aborted status
	SetAborted(bool)
	// IsAborted checks the event's aborted status
	IsAborted() bool
}

// BaseEvent serves as a basic implementation of the `Event` interface and contains fields for storing the topic,
// payload, and aborted status of an event. The struct includes methods to interact with these fields
// such as getting and setting the payload, setting the aborted status, and checking if the event has
// been aborted. The struct also includes a `sync.RWMutex` field `mu` to handle concurrent access to
// the struct's fields in a thread-safe manner
type BaseEvent struct {
	topic      string
	payload    interface{}
	aborted    bool
	properties Properties
	mu         sync.RWMutex
}

// NewBaseEvent creates a new instance of BaseEvent with a payload
func NewBaseEvent(topic string, payload interface{}) *BaseEvent {
	return &BaseEvent{
		topic:   topic,
		payload: payload,
	}
}

// Topic returns the event's topic
func (e *BaseEvent) Topic() string {
	return e.topic
}

// Payload returns the event's payload
func (e *BaseEvent) Payload() interface{} {
	e.mu.RLock() // Read lock
	defer e.mu.RUnlock()

	return e.payload
}

// SetPayload sets the event's payload
func (e *BaseEvent) SetPayload(payload interface{}) {
	e.mu.Lock() // Write lock
	defer e.mu.Unlock()
	e.payload = payload
}

// Properties returns the event's properties
func (e *BaseEvent) Properties() Properties {
	e.mu.RLock() // Read lock
	defer e.mu.RUnlock()

	return e.properties
}

// SetProperties sets the event's properties
func (e *BaseEvent) SetProperties(properties Properties) {
	e.mu.Lock() // Write lock
	defer e.mu.Unlock()
	e.properties = properties
}

// SetAborted sets the event's aborted status
func (e *BaseEvent) SetAborted(abort bool) {
	e.mu.Lock() // Write lock
	defer e.mu.Unlock()
	e.aborted = abort
}

// IsAborted checks the event's aborted status
func (e *BaseEvent) IsAborted() bool {
	e.mu.RLock() // Read lock
	defer e.mu.RUnlock()

	return e.aborted
}

// Properties is a map of properties to set on an event
type Properties map[string]interface{}

// NewProperties creates a new Properties map
func NewProperties() Properties {
	return make(Properties, 10) //nolint:mnd
}

// Set a property on the Properties map
func (p Properties) Set(name string, value interface{}) Properties {
	p[name] = value
	return p
}
