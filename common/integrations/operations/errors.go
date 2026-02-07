package operations

import "errors"

var (
	// ErrDecodeConfigTargetNil indicates DecodeConfig was called with a nil target
	ErrDecodeConfigTargetNil = errors.New("operations: decode config target is nil")
	// ErrOperationTemplateRequired indicates the operation requires a stored template configuration
	ErrOperationTemplateRequired = errors.New("operations: operation template required")
	// ErrOperationTemplateOverridesNotAllowed indicates overrides are not permitted for a template
	ErrOperationTemplateOverridesNotAllowed = errors.New("operations: operation template overrides not allowed")
	// ErrOperationTemplateOverrideNotAllowed indicates a provided override key is not permitted
	ErrOperationTemplateOverrideNotAllowed = errors.New("operations: operation template override not allowed")
)
