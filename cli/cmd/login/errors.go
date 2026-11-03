//go:build cli

package login

import "errors"

// ErrNoSessionIssued is returned when 2FA validation succeeds but the server did not issue a session
var ErrNoSessionIssued = errors.New("2FA validation succeeded but the server did not issue a session")
