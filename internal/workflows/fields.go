package workflows

import (
	"github.com/theopenlane/core/common/enums"
)

// EligibleWorkflowFields returns the canonical set of workflow-eligible fields for an object type.
func EligibleWorkflowFields(objectType enums.WorkflowObjectType) map[string]struct{} {
	schema, err := WorkflowSchema(objectType)
	if err != nil {
		return map[string]struct{}{}
	}

	fields := schema.WorkflowFields()
	eligible := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		eligible[field.Name] = struct{}{}
	}

	return eligible
}
