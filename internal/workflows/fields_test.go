package workflows

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/common/enums"
)

func TestEligibleWorkflowFields(t *testing.T) {
	metadata := WorkflowMetadata()
	assert.NotEmpty(t, metadata)

	entry := metadata[0]
	fields := EligibleWorkflowFields(entry.Type)
	assert.NotEmpty(t, fields)

	for _, field := range entry.EligibleFields {
		assert.Contains(t, fields, field.Name)
	}

	unknown := EligibleWorkflowFields(enums.WorkflowObjectType("Unknown"))
	assert.Empty(t, unknown)
}
