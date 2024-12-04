package soiree

import "github.com/cenkalti/backoff/v4"

// Listener is a function type that can handle events of any type
// Listener takes an `Event` as a parameter and returns an `error`. This
// allows you to define functions that conform to this specific signature, making it easier to work
// with event listeners in the other parts of the code
type Listener func(Event) error

// listenerItem stores a listener along with its unique identifier and priority
type listenerItem struct {
	listener Listener
	priority Priority
	client   interface{}
	retries  int
	backoff  backoff.BackOff
}

// ListenerOption is a function type that configures listener behavior
type ListenerOption func(*listenerItem)

// WithPriority sets the priority of a listener
func WithPriority(priority Priority) ListenerOption {
	return func(item *listenerItem) {
		item.priority = priority
	}
}

// WithClient sets the client of a listener
func WithListenerClient(client interface{}) ListenerOption {
	return func(item *listenerItem) {
		item.client = client
	}
}

func WithRetry(retry int) ListenerOption {
	return func(item *listenerItem) {
		item.client = retry
	}
}

func WithBackoff(backoff backoff.BackOff) ListenerOption {
	return func(item *listenerItem) {
		item.backoff = backoff
	}
}
