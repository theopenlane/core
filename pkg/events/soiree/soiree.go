package soiree

// Soiree is an interface that defines the behavior of your get-together
type Soiree interface {
	// On registers a listener function to a specific topic
	On(topicName string, listener Listener, opts ...ListenerOption) (string, error)
	// Off removes a listener from a specific topic using the listener's unique ID
	Off(topicName string, listenerID string) error
	// Emit asynchronously sends an event to all subscribers of a topic and returns a channel of errors
	Emit(eventName string, payload interface{}) <-chan error
	// EmitSync sends an event synchronously to all subscribers of a topic; blocks until all listeners have been notified
	EmitSync(eventName string, payload interface{}) []error
	// GetTopic retrieves the Topic object associated with the given topic name
	GetTopic(topicName string) (*Topic, error)
	// EnsureTopic creates a new topic if it does not exist, or returns the existing one
	EnsureTopic(topicName string) *Topic
	// SetErrorHandler assigns a custom error handler function for the Soiree
	SetErrorHandler(func(Event, error) error)
	// SetIDGenerator assigns a function that generates a unique ID string for new listeners
	SetIDGenerator(func() string)
	// SetPool sets a custom goroutine pool for managing concurrency within the Soiree
	SetPool(Pool)
	// SetPanicHandler sets a function that will be called in case of a panic during event handling
	SetPanicHandler(PanicHandler)
	// SetErrChanBufferSize sets the size of the buffered channel for errors returned by asynchronous emits
	SetErrChanBufferSize(int)
	// Close gracefully shuts down the Soiree, ensuring all pending events are processed
	Close() error
	// SetClient sets the client for the Soiree
	SetClient(client interface{})
	// GetClient gets the client for the Soiree
	GetClient() interface{}
}
