package buildkite

import "errors"

var (
	// ErrAPITokenMissing indicates the Buildkite API token is missing from the credential
	ErrAPITokenMissing = errors.New("buildkite: api token missing")
	// ErrClientType indicates the provided client is not a Buildkite client
	ErrClientType = errors.New("buildkite: unexpected client type")
)
