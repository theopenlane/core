package workflows

import (
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/mapx"
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

// SeparateFieldsByEligibility splits fields into eligible (workflow-controlled) and
// ineligible (pass-through) sets for a given schema type.
func SeparateFieldsByEligibility(schemaType string, fields []string) (eligible, ineligible []string) {
	schema, ok := entityops.LookupSchema(schemaType)
	if !ok || !schema.WorkflowEligible {
		return nil, fields
	}

	eligibleSet := make(map[string]struct{})
	for _, field := range schema.WorkflowFields() {
		eligibleSet[field.Name] = struct{}{}
	}

	for _, field := range fields {
		if _, ok := eligibleSet[field]; ok {
			eligible = append(eligible, field)
		} else {
			ineligible = append(ineligible, field)
		}
	}

	return eligible, ineligible
}

// CollectChangedFields returns the union of modified and cleared fields from a mutation,
// filtered to only include fields eligible for workflow processing.
func CollectChangedFields(m utils.GenericMutation) []string {
	uniqueFields := changedAndClearedFields(m)
	schema, ok := entityops.LookupSchema(m.Type())
	if !ok || !schema.WorkflowEligible {
		return uniqueFields
	}

	eligible := make(map[string]struct{})
	for _, field := range schema.WorkflowFields() {
		eligible[field.Name] = struct{}{}
	}

	return lo.Filter(uniqueFields, func(f string, _ int) bool {
		_, isEligible := eligible[f]
		return isEligible
	})
}

// CollectAllChangedFields returns the union of modified and cleared fields from a mutation
// without filtering by workflow eligibility.
func CollectAllChangedFields(m utils.GenericMutation) []string {
	return changedAndClearedFields(m)
}

// changedAndClearedFields returns the normalized union of modified and cleared fields
func changedAndClearedFields(m utils.GenericMutation) []string {
	return NormalizeStrings(append(append([]string(nil), m.Fields()...), m.ClearedFields()...))
}

// NormalizeStrings trims, deduplicates, and drops empty string values
func NormalizeStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	normalized := lo.Uniq(lo.FilterMap(values, func(value string, _ int) (string, bool) {
		value = strings.TrimSpace(value)
		return value, value != ""
	}))
	if len(normalized) == 0 {
		return nil
	}

	return normalized
}

// BuildProposedChanges materializes mutation values including explicit clears for the
// pre-commit approval routing path
func BuildProposedChanges(m utils.GenericMutation, changedFields []string) map[string]any {
	if m == nil || len(changedFields) == 0 {
		return nil
	}

	clearedSet := mapx.MapSetFromSlice(NormalizeStrings(m.ClearedFields()))

	proposed := make(map[string]any, len(changedFields))
	for _, field := range changedFields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}

		if value, ok := m.Field(field); ok {
			proposed[field] = value
			continue
		}

		if _, ok := clearedSet[field]; ok {
			proposed[field] = nil
		}
	}

	if len(proposed) == 0 {
		return nil
	}

	return proposed
}
