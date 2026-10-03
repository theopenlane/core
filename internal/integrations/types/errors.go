package types //nolint:revive

import "errors"

var (
	// ErrClientCastFailed indicates a registered client could not be cast to the expected type
	ErrClientCastFailed = errors.New("integrations: client cast failed")
	// ErrCredentialRefNotFound indicates the requested credential ref was not found in the definition
	ErrCredentialRefNotFound = errors.New("integrations: credential ref not found")
	// ErrConnectionRefNotFound indicates the connection ref was not found in the definition
	ErrConnectionRefNotFound = errors.New("integrations: connection credential ref not found")
	// ErrOperationConfigInvalid indicates an operation config payload could not be decoded
	ErrOperationConfigInvalid = errors.New("operation config invalid")
)

// UnhealthyError marks an operation failure as terminal, requiring user action to recover
type UnhealthyError struct {
	// Reason is the user-facing explanation included in the organization notification
	Reason string
	// Err is the underlying failure
	Err error
}

// Error returns the reason followed by the underlying failure
func (e *UnhealthyError) Error() string {
	return e.Reason + ": " + e.Err.Error()
}

// Unwrap exposes the underlying failure for errors.Is and errors.As
func (e *UnhealthyError) Unwrap() error {
	return e.Err
}

// Unhealthy wraps err as a terminal integration failure with a user-facing reason
func Unhealthy(err error, reason string) error {
	return &UnhealthyError{Reason: reason, Err: err}
}

// UnhealthyFrom returns the UnhealthyError in err's chain when present
func UnhealthyFrom(err error) (*UnhealthyError, bool) {
	return errors.AsType[*UnhealthyError](err)
}

// DegradedError marks a single operation's failure as terminal while the rest keeps running
type DegradedError struct {
	// Reason is the user-facing explanation recorded against the operation
	Reason string
	// Err is the underlying failure
	Err error
}

// Error returns the reason followed by the underlying failure
func (e *DegradedError) Error() string {
	return e.Reason + ": " + e.Err.Error()
}

// Unwrap exposes the underlying failure for errors.Is and errors.As
func (e *DegradedError) Unwrap() error {
	return e.Err
}

// Degraded wraps err as a terminal failure for the executing operation with a user-facing reason
func Degraded(err error, reason string) error {
	return &DegradedError{Reason: reason, Err: err}
}

// DegradedFrom returns the DegradedError in err's chain when present
func DegradedFrom(err error) (*DegradedError, bool) {
	return errors.AsType[*DegradedError](err)
}
