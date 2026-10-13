//go:build examples

package integration

import "errors"

var (
	// ErrDefinitionIDRequired is returned when --definition-id is missing
	ErrDefinitionIDRequired = errors.New("--definition-id is required")
	// ErrInvalidBody is returned when --body can't be parsed as JSON
	ErrInvalidBody = errors.New("--body must be valid JSON or @path/to/file.json")
	// ErrInvalidUserInput is returned when --user-input can't be parsed as JSON
	ErrInvalidUserInput = errors.New("--user-input must be valid JSON or @path/to/file.json")
	// ErrInvalidOperationConfig is returned when --operation-config can't be parsed as a JSON object keyed by operation name
	ErrInvalidOperationConfig = errors.New("--operation-config must be a JSON object keyed by operation name or @path/to/file.json")
	// ErrInvalidJSON is returned when a JSON flag contains malformed JSON
	ErrInvalidJSON = errors.New("invalid JSON")
)
