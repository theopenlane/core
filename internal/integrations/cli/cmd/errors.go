//go:build examples

package cmd

import "errors"

// ErrLivezUnhealthy is returned when the livez endpoint returns a non-200 status
var ErrLivezUnhealthy = errors.New("livez returned unhealthy status")
