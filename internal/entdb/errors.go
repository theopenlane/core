package entdb

import (
	"errors"

	"github.com/theopenlane/core/internal/shutdown"
)

var (
	// ErrDriverLackingBeginTx is returned when the driver does not support BeginTx
	ErrDriverLackingBeginTx = errors.New("driver does not support BeginTx")
	// ErrShuttingDown is returned when operations are attempted during shutdown
	ErrShuttingDown = shutdown.ErrShuttingDown
)
