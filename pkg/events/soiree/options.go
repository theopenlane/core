package soiree

import (
	"github.com/cenkalti/backoff/v5"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/utils/ulids"
)

// EventPoolOption defines a function type for Soiree configuration options
type EventPoolOption func(Soiree)

var DefaultErrorHandler = func(_ Event, err error) error {
	return err
}

// DefaultIDGenerator generates a unique identifier
var DefaultIDGenerator = func() string {
	return ulids.New().String()
}

// DefaultPanicHandler handles panics by printing the panic value
var DefaultPanicHandler = func(p any) {
	log.Error().Msgf("panic occurred processing event: %v", p)
}

// WithErrorHandler sets a custom error handler for an Soiree
func WithErrorHandler(errHandler func(Event, error) error) EventPoolOption {
	return func(m Soiree) {
		m.SetErrorHandler(errHandler)
	}
}

// WithIDGenerator sets a custom ID generator for an Soiree
func WithIDGenerator(idGen func() string) EventPoolOption {
	return func(m Soiree) {
		m.SetIDGenerator(idGen)
	}
}

// WithPool sets a custom pool for an Soiree
func WithPool(pool Pool) EventPoolOption {
	return func(m Soiree) {
		m.SetPool(pool)
	}
}

// WithEventStore sets the event persistence store for the Soiree
func WithEventStore(store EventStore) EventPoolOption {
	return func(m Soiree) {
		m.SetEventStore(store)
	}
}

// WithRedisStore configures a Redis-backed event store
func WithRedisStore(client *redis.Client) EventPoolOption {
	return func(m Soiree) {
		m.SetEventStore(NewRedisStore(client))
	}
}

// WithRetry configures retry attempts and backoff behavior for listener failures
func WithRetry(retries int, factory func() backoff.BackOff) EventPoolOption {
	return func(m Soiree) {
		m.SetRetry(retries, factory)
	}
}

// PanicHandler is a function type that handles panics
type PanicHandler func(any)

// WithPanicHandler sets a custom panic handler for an Soiree
func WithPanicHandler(panicHandler PanicHandler) EventPoolOption {
	return func(m Soiree) {
		m.SetPanicHandler(panicHandler)
	}
}

// WithErrChanBufferSize sets the size of the buffered channel for errors returned by asynchronous emits
func WithErrChanBufferSize(size int) EventPoolOption {
	return func(m Soiree) {
		m.SetErrChanBufferSize(size)
	}
}

// WithClient sets a custom client for the Soiree
func WithClient(client any) EventPoolOption {
	return func(m Soiree) {
		m.SetClient(client)
	}
}
