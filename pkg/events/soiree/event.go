package soiree

import (
	"context"
	"sync"
)

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
	// Context returns the event's context
	Context() context.Context
	// SetContext sets the event's context
	SetContext(context.Context)
	// Client returns the event's client
	Client() interface{}
	// SetClient sets the event's client
	SetClient(interface{})
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
	ctx        context.Context
	client     interface{}
}

// NewBaseEvent creates a new instance of BaseEvent with a payload
func NewBaseEvent(topic string, payload interface{}) *BaseEvent {
	return &BaseEvent{
		topic:      topic,
		payload:    payload,
		properties: Properties{},
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
	if properties == nil {
		properties = NewProperties()
	}

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
	return make(Properties)
}

// Set a property on the Properties map
func (p Properties) Set(name string, value interface{}) Properties {
	if p == nil {
		p = NewProperties()
	}

	p[name] = value

	return p
}

// Get a property from the Properties map
func (p Properties) GetKey(key string) interface{} {
	value := p[key]

	if value == nil || value == "" {
		return nil
	}

	return value
}

// Context returns the event's context
func (e *BaseEvent) Context() context.Context {
	e.mu.RLock() // Read lock
	defer e.mu.RUnlock()

	return e.ctx
}

// SetContext sets the event's context
func (e *BaseEvent) SetContext(ctx context.Context) {
	e.mu.Lock() // Write lock
	defer e.mu.Unlock()
	e.ctx = ctx
}

// Client returns the event's client
func (e *BaseEvent) Client() interface{} {
	e.mu.RLock() // Read lock
	defer e.mu.RUnlock()

	return e.client
}

// SetClient sets the event's client
func (e *BaseEvent) SetClient(client interface{}) {
	e.mu.Lock() // Write lock
	defer e.mu.Unlock()

	e.client = client
}
