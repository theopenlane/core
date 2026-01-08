package soiree

import (
	"github.com/cenkalti/backoff/v5"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/utils/ulids"
)

// Option defines a function type for EventBus configuration options
type Option func(*EventBus)

// defaultErrorHandler is the default error handler that simply returns the error
var defaultErrorHandler = func(_ Event, err error) error {
	return err
}

// defaultIDGenerator generates a unique identifier
var defaultIDGenerator = func() string {
	return ulids.New().String()
}

// defaultPanicHandler handles panics by printing the panic value
var defaultPanicHandler = func(p any) {
	log.Error().Msgf("panic occurred processing event: %v", p)
}

// ErrorHandler sets a custom error handler for an EventBus
func ErrorHandler(errHandler func(Event, error) error) Option {
	return func(m *EventBus) {
		m.errorHandler = errHandler
	}
}

// IDGenerator sets a custom ID generator for an EventBus
func IDGenerator(idGen func() string) Option {
	return func(m *EventBus) {
		m.idGenerator = idGen
	}
}

// Workers sets the number of worker goroutines for concurrent event handling
func Workers(n int) Option {
	return func(m *EventBus) {
		if n > 0 {
			m.poolOpts = append(m.poolOpts, WithWorkers(n))
		}
	}
}

// EventStore configures a custom event store
func EventStore(store eventStore) Option {
	return func(m *EventBus) {
		m.store = store
	}
}

// WithRedisStore configures a Redis-backed event store
func WithRedisStore(client *redis.Client, opts ...RedisStoreOption) Option {
	return func(m *EventBus) {
		m.store = NewRedisStore(client, opts...)
	}
}

// Retry configures retry attempts and backoff behavior for listener failures
func Retry(retries int, factory func() backoff.BackOff) Option {
	return func(m *EventBus) {
		if retries > 0 {
			m.maxRetries = retries
		}
		if factory != nil {
			m.backOffFactory = factory
		}
	}
}

// PanicHandler is a function type that handles panics
type PanicHandler func(any)

// Panics sets a custom panic handler for an EventBus
func Panics(panicHandler PanicHandler) Option {
	return func(m *EventBus) {
		m.panicHandler = panicHandler
	}
}

// ErrChanBufferSize sets the size of the buffered channel for errors returned by asynchronous emits
func ErrChanBufferSize(size int) Option {
	return func(m *EventBus) {
		if size < 1 {
			size = 1
		}
		m.errChanBufferSize = size
	}
}

// Client sets a custom client for the EventBus
func Client(client any) Option {
	return func(m *EventBus) {
		m.client = client
	}
}
