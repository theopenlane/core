package types //nolint:revive

import "errors"

var (
	// ErrClientCastFailed indicates a registered client instance could not be cast to the expected type
	ErrClientCastFailed = errors.New("integrations: client cast failed")
	// ErrCredentialRefNotFound indicates the requested credential ref was not found in the definition
	ErrCredentialRefNotFound = errors.New("integrations: credential ref not found")
	// ErrConnectionRefNotFound indicates the requested connection credential ref was not found in the definition
	ErrConnectionRefNotFound = errors.New("integrations: connection credential ref not found")
)

// UnhealthyError marks an operation failure as terminal for recurring cycles: recovery
// requires user action, so the runtime stops the loop instead of retrying forever
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
