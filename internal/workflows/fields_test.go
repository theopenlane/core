package workflows

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/common/enums"
)

func TestEligibleWorkflowFields(t *testing.T) {
	metadata := WorkflowMetadata()
	assert.NotEmpty(t, metadata)

	// Register eligible fields from metadata for this test
	fieldsMap := make(map[enums.WorkflowObjectType]map[string]struct{})
	for _, info := range metadata {
		eligible := make(map[string]struct{})
		for _, f := range info.EligibleFields {
			eligible[f.Name] = struct{}{}
		}
		fieldsMap[info.Type] = eligible
	}
	RegisterEligibleFields(fieldsMap)
	t.Cleanup(func() { RegisterEligibleFields(nil) })

	entry := metadata[0]
	fields := EligibleWorkflowFields(entry.Type)
	assert.NotEmpty(t, fields)

	for _, field := range entry.EligibleFields {
		assert.Contains(t, fields, field.Name)
	}

	unknown := EligibleWorkflowFields(enums.WorkflowObjectType("Unknown"))
	assert.Empty(t, unknown)
}

func TestCollectChangedFields(t *testing.T) {
	metadata := WorkflowMetadata()
	assert.NotEmpty(t, metadata)
	assert.NotEmpty(t, metadata[0].EligibleFields)

	// Register eligible fields from metadata for this test
	fieldsMap := make(map[enums.WorkflowObjectType]map[string]struct{})
	for _, info := range metadata {
		eligible := make(map[string]struct{})
		for _, f := range info.EligibleFields {
			eligible[f.Name] = struct{}{}
		}
		fieldsMap[info.Type] = eligible
	}
	RegisterEligibleFields(fieldsMap)
	t.Cleanup(func() { RegisterEligibleFields(nil) })

	eligibleName := metadata[0].EligibleFields[0].Name
	m := fakeMutation{
		typ:     metadata[0].Type.String(),
		fields:  []string{eligibleName, "ignore", eligibleName},
		cleared: []string{"ignore2"},
		values: map[string]any{
			eligibleName: "value",
		},
	}

	changed := CollectChangedFields(m)
	assert.ElementsMatch(t, []string{eligibleName}, changed)

	// For unknown types, CollectChangedFields returns all unique fields unfiltered
	m.typ = "UnknownType"
	changed = CollectChangedFields(m)
	assert.ElementsMatch(t, []string{eligibleName, "ignore", "ignore2"}, changed)
}

func TestSeparateFieldsByEligibility(t *testing.T) {
	metadata := WorkflowMetadata()
	assert.NotEmpty(t, metadata)
	assert.NotEmpty(t, metadata[0].EligibleFields)

	fieldsMap := make(map[enums.WorkflowObjectType]map[string]struct{})
	for _, info := range metadata {
		eligible := make(map[string]struct{})
		for _, f := range info.EligibleFields {
			eligible[f.Name] = struct{}{}
		}
		fieldsMap[info.Type] = eligible
	}
	RegisterEligibleFields(fieldsMap)
	t.Cleanup(func() { RegisterEligibleFields(nil) })

	eligibleName := metadata[0].EligibleFields[0].Name
	fields := []string{eligibleName, "ineligible"}

	eligible, ineligible := SeparateFieldsByEligibility(metadata[0].Type.String(), fields)
	assert.ElementsMatch(t, []string{eligibleName}, eligible)
	assert.ElementsMatch(t, []string{"ineligible"}, ineligible)

	eligible, ineligible = SeparateFieldsByEligibility("UnknownType", fields)
	assert.Empty(t, eligible)
	assert.ElementsMatch(t, fields, ineligible)
}
