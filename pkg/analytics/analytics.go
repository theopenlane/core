package analytics

import (
	ph "github.com/posthog/posthog-go"
)

// EventManager isn't your normal party planner
type EventManager struct {
	Enabled bool
	Handler Handler
}

// Handler is an interface which can be used to call various event / event association parameters provided by the posthog API
type Handler interface {
	Event(eventName string, properties ph.Properties)
}

// Event function is used to send an event to the analytics handler
func (e *EventManager) Event(eventName string, properties ph.Properties) {
	if e.Enabled {
		e.Handler.Event(eventName, properties)
	}
}
