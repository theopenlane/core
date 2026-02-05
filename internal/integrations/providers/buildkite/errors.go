package buildkite

import "errors"

var (
	// ErrAPIRequest indicates a Buildkite API request failed with a non-2xx status
	ErrAPIRequest = errors.New("buildkite: api request failed")
)
