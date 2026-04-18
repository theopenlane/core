package emailtemplate

import "errors"

var (
	// ErrKeyRequired is returned when --key is missing
	ErrKeyRequired = errors.New("--key is required")
	// ErrNameRequired is returned when --name is missing
	ErrNameRequired = errors.New("--name is required")
	// ErrTemplateContextRequired is returned when --template-context is missing
	ErrTemplateContextRequired = errors.New("--template-context is required")
)
