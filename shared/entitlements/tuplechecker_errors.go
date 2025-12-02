package entitlements

import "errors"

var (
	// ErrTupleCheckerNotConfigured is returned when TupleChecker is not properly configured
	ErrTupleCheckerNotConfigured = errors.New("TupleChecker not properly configured")
)
