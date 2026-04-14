package models

import (
	"encoding/json"

	"github.com/theopenlane/core/common/enums"
)

// TemplateVariable describes a single system-provided template variable available
// for use in email templates
type TemplateVariable struct {
	// Name is the variable key as used in templates (e.g. "companyName" for {{ .companyName }})
	Name string `json:"name"`
	// Description explains what the variable contains
	Description string `json:"description"`
}

// TemplateContextEntry describes a registered template data context, its human-readable
// label and description, and the available template variables for UI tooling
type TemplateContextEntry struct {
	// Context is the TemplateContext enum value identifying this context
	Context enums.TemplateContext `json:"context"`
	// Label is a human-readable name for this context
	Label string `json:"label"`
	// Description explains when this context is used
	Description string `json:"description"`
	// Schema is the JSON Schema for this context's template data shape, reflected at init time
	Schema json.RawMessage `json:"schema"`
	// ReservedFields lists top-level template variable names injected by the system at
	// render time. These are available in templates but are not user-supplied inputs
	ReservedFields []string `json:"reservedFields"`
	// Variables lists all system-provided template variables available in this context,
	// with human-readable descriptions for the UI variable picker
	Variables []TemplateVariable `json:"variables"`
}
