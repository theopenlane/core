package models

import "github.com/theopenlane/core/common/enums"

// TemplateContextEntry describes a registered template data context, its human-readable
// label and description, and the reflected JSON Schema for UI tooling.
type TemplateContextEntry struct {
	// Context is the TemplateContext enum value identifying this context.
	Context enums.TemplateContext `json:"context"`
	// Label is a human-readable name for this context.
	Label string `json:"label"`
	// Description explains when this context is used.
	Description string `json:"description"`
	// Schema is the JSON Schema for this context's template data shape, reflected at init time.
	// Intended for UI tooling only — not used for runtime validation.
	Schema map[string]any `json:"schema"`
}
