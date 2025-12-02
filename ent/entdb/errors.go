package entdb

import (
	"errors"
)

var (
	// ErrDriverLackingBeginTx is returned when the driver does not support BeginTx
	ErrDriverLackingBeginTx = errors.New("driver does not support BeginTx")
	// ErrShuttingDown is returned when operations are attempted during shutdown
	ErrShuttingDown = errors.New("database shutting down")
)
