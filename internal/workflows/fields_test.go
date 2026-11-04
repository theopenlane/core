package workflows

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
)

func TestEligibleWorkflowFields(t *testing.T) {
	metadata := WorkflowMetadata()
	assert.Check(t, len(metadata) != 0)

	entry := metadata[0]
	fields := EligibleWorkflowFields(entry.Type)
	assert.Check(t, len(fields) != 0)

	for _, field := range entry.EligibleFields {
		assert.Check(t, is.Contains(fields, field.Name))
	}

	unknown := EligibleWorkflowFields(enums.WorkflowObjectType("Unknown"))
	assert.Check(t, is.Len(unknown, 0))
}
