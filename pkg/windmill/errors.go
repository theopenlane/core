//go:build windmill

package windmill

import "errors"

var (
	// ErrWindmillDisabled is returned when Windmill integration is disabled
	ErrWindmillDisabled = errors.New("windmill integration is disabled")

	// ErrMissingToken is returned when no API token is configured
	ErrMissingToken = errors.New("windmill API token is required")

	// ErrMissingWorkspace is returned when no workspace is configured
	ErrMissingWorkspace = errors.New("windmill workspace is required")

	// ErrInvalidFlowPath is returned when the flow path is invalid
	ErrInvalidFlowPath = errors.New("invalid flow path")

	// ErrFlowNotFound is returned when a flow is not found
	ErrFlowNotFound = errors.New("flow not found")
)
