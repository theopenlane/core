package workflows

import (
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
)

// eligibleFieldsRegistry holds the registered eligible fields map from generated code.
var eligibleFieldsRegistry map[enums.WorkflowObjectType]map[string]struct{}

// RegisterEligibleFields sets the eligible fields registry from generated code.
func RegisterEligibleFields(fields map[enums.WorkflowObjectType]map[string]struct{}) {
	eligibleFieldsRegistry = fields
}

// EligibleWorkflowFields returns the set of fields eligible for workflow processing for a given object type.
func EligibleWorkflowFields(objectType enums.WorkflowObjectType) map[string]struct{} {
	if eligibleFieldsRegistry == nil {
		return map[string]struct{}{}
	}

	fields, ok := eligibleFieldsRegistry[objectType]
	if !ok {
		return map[string]struct{}{}
	}

	return fields
}

// CollectChangedFields returns the union of modified and cleared fields from a mutation,
// filtered to only include fields eligible for workflow processing.
func CollectChangedFields(m utils.GenericMutation) []string {
	fields := m.Fields()
	cleared := m.ClearedFields()

	objectType := enums.ToWorkflowObjectType(m.Type())
	eligible := EligibleWorkflowFields(lo.FromPtr(objectType))

	allFields := append(append([]string(nil), fields...), cleared...)
	uniqueFields := lo.Uniq(allFields)

	if len(eligible) == 0 {
		return uniqueFields
	}

	return lo.Filter(uniqueFields, func(f string, _ int) bool {
		_, isEligible := eligible[f]
		return isEligible
	})
}

// CollectAllChangedFields returns the union of modified and cleared fields from a mutation
// without filtering by workflow eligibility.
func CollectAllChangedFields(m utils.GenericMutation) []string {
	fields := m.Fields()
	cleared := m.ClearedFields()

	allFields := append(append([]string(nil), fields...), cleared...)
	return lo.Uniq(allFields)
}
