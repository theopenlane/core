package workflows

import (
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
)

// EligibleWorkflowFields returns the set of fields eligible for workflow processing for a given object type.
func EligibleWorkflowFields(objectType enums.WorkflowObjectType) map[string]struct{} {
	entry, found := lo.Find(WorkflowMetadata(), func(e generated.WorkflowObjectTypeInfo) bool {
		return e.Type == objectType
	})
	if !found {
		return map[string]struct{}{}
	}

	return lo.SliceToMap(entry.EligibleFields, func(f generated.WorkflowFieldInfo) (string, struct{}) {
		return f.Name, struct{}{}
	})
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

	return lo.Filter(uniqueFields, func(f string, _ int) bool {
		_, isEligible := eligible[f]
		return isEligible
	})
}
