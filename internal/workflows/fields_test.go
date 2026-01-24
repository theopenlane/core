package workflows

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/enums"
)

func TestEligibleWorkflowFields(t *testing.T) {
	metadata := WorkflowMetadata()
	require.NotEmpty(t, metadata)

	entry := metadata[0]
	fields := EligibleWorkflowFields(entry.Type)
	require.NotEmpty(t, fields)

	for _, field := range entry.EligibleFields {
		assert.Contains(t, fields, field.Name)
	}

	unknown := EligibleWorkflowFields(enums.WorkflowObjectType("Unknown"))
	assert.Empty(t, unknown)
}

func TestCollectChangedFields(t *testing.T) {
	metadata := WorkflowMetadata()
	require.NotEmpty(t, metadata)
	require.NotEmpty(t, metadata[0].EligibleFields)

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

	m.typ = "UnknownType"
	changed = CollectChangedFields(m)
	assert.Empty(t, changed)
}
