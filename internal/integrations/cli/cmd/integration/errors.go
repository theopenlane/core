package integration

import "errors"

var (
	// ErrDefinitionIDRequired is returned when --definition-id is missing
	ErrDefinitionIDRequired = errors.New("--definition-id is required")
	// ErrInvalidBody is returned when --body can't be parsed as JSON
	ErrInvalidBody = errors.New("--body must be valid JSON or @path/to/file.json")
	// ErrInvalidUserInput is returned when --user-input can't be parsed as JSON
	ErrInvalidUserInput = errors.New("--user-input must be valid JSON or @path/to/file.json")
	// ErrInvalidJSON is returned when a JSON flag contains malformed JSON
	ErrInvalidJSON = errors.New("invalid JSON")
)
