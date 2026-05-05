//go:build examples

package cmd

import "errors"

var (
	// ErrUnsupportedOutputFormat is returned when an output format is not supported
	ErrUnsupportedOutputFormat = errors.New("unsupported output format")
	// ErrLivezUnhealthy is returned when the livez endpoint returns a non-200 status
	ErrLivezUnhealthy = errors.New("livez returned unhealthy status")
)
