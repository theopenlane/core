package soiree

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
}

// ListenerOption is a function type that configures listener behavior
type ListenerOption func(*listenerItem)

// WithPriority sets the priority of a listener
func WithPriority(priority Priority) ListenerOption {
	return func(item *listenerItem) {
		item.priority = priority
	}
}
