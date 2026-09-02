//go:build test

package integrations

import "errors"

var (
	// ErrCycleFailed is returned by the always-failing reconcile operation
	ErrCycleFailed = errors.New("integrations: cycle failed")
	// ErrTokenMissing indicates no usable token credential is stored for the installation
	ErrTokenMissing = errors.New("integrations: token missing")
	// ErrHealthFailed is returned by the health check when a failure marker credential is bound
	ErrHealthFailed = errors.New("integrations: health failed")
	// ErrOAuthCodeMissing indicates the OAuth callback carried no code
	ErrOAuthCodeMissing = errors.New("integrations: missing oauth code")
	// ErrOAuthStateMismatch indicates the OAuth callback state did not match
	ErrOAuthStateMismatch = errors.New("integrations: oauth state mismatch")
)
