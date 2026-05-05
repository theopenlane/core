//go:build examples

package emailtemplate

import "errors"

var (
	// ErrKeyRequired is returned when --key is missing
	ErrKeyRequired = errors.New("--key is required")
	// ErrNameRequired is returned when --name is missing
	ErrNameRequired = errors.New("--name is required")
	// ErrTemplateContextRequired is returned when --template-context is missing
	ErrTemplateContextRequired = errors.New("--template-context is required")
	// ErrDefaultsFileUnreadable is returned when the file referenced by --defaults-file cannot be read
	ErrDefaultsFileUnreadable = errors.New("--defaults-file could not be read")
	// ErrDefaultsInvalid is returned when defaults input cannot be parsed as a JSON object
	ErrDefaultsInvalid = errors.New("defaults must be a valid JSON object")
)
