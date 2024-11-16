package soiree

import (
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/utils/ulids"
)

// EventPoolOption defines a function type for Soiree configuration options
type EventPoolOption func(Soiree)

var DefaultErrorHandler = func(event Event, err error) error {
	return err
}

// DefaultIDGenerator generates a unique identifier
var DefaultIDGenerator = func() string {
	return ulids.New().String()
}

// DefaultPanicHandler handles panics by printing the panic value
var DefaultPanicHandler = func(p interface{}) {
	log.Error().Msgf("Panic occurred: %v", p)
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

// PanicHandler is a function type that handles panics
type PanicHandler func(interface{})

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

func WithClient(client interface{}) EventPoolOption {
	return func(m Soiree) {
		m.SetClient(client)
	}
}
