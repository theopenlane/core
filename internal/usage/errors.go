package usage

import "errors"

var (
	// ErrUsageLimitReached is returned when a usage limit would be exceeded.
	ErrUsageLimitReached = errors.New("usage limit reached")
)
