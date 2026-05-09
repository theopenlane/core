package templatekit

import "errors"

var (
	// ErrTemplateNotFound is returned when a notification template cannot be resolved
	ErrTemplateNotFound = errors.New("notification template not found")
	// ErrTemplateRenderFailed is returned when template payload merging fails
	ErrTemplateRenderFailed = errors.New("template render failed")
)
