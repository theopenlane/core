// Package shutdown holds the process-wide shutdown flag
package shutdown

import (
	"errors"
	"sync/atomic"
)

// ErrShuttingDown is returned when operations are attempted during shutdown
var ErrShuttingDown = errors.New("database client shutting down")

// Flag tracks whether a shutdown is in progress
type Flag struct {
	// flag is an atomic boolean that indicates if a shutdown is in progress
	// atomic.Bool is used to ensure thread-safe access to the flag
	// without the need for explicit locks
	// this is important in a concurrent environment where multiple goroutines
	// may be checking or setting the flag at the same time
	flag atomic.Bool
}

// Begin sets the shutdown flag to true, indicating that a shutdown is in progress
func (s *Flag) Begin() {
	s.flag.Store(true)
}

// Reset clears the shutdown flag, indicating that a shutdown is no longer in progress
func (s *Flag) Reset() {
	s.flag.Store(false)
}

// IsSet checks if the shutdown flag is set to true
func (s *Flag) IsSet() bool {
	return s.flag.Load()
}

// New creates a new instance of Flag
func New() *Flag {
	return &Flag{}
}

// Default is the shutdown flag used throughout the application
var Default = New()

// Begin marks the system as shutting down. It is safe to call multiple times
func Begin() {
	Default.Begin()
}

// Reset clears the shutdown flag. It is intended for tests
func Reset() {
	Default.Reset()
}

// InProgress reports whether shutdown has begun
func InProgress() bool {
	return Default.IsSet()
}

// IsError reports whether the error came from an operation refused because the process is
// shutting down. These are expected during a deploy or restart and should be retried rather
// than logged as failures
func IsError(err error) bool {
	return errors.Is(err, ErrShuttingDown)
}
