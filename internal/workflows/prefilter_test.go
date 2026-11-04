package workflows

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/models"
)

// TestDeriveTriggerPrefilter verifies trigger prefilter derivation
func TestDeriveTriggerPrefilter(t *testing.T) {
	doc := models.WorkflowDefinitionDocument{
		Triggers: []models.WorkflowTrigger{
			{
				Operation: "update",
				Fields:    []string{"status", ""},
				Edges:     []string{"evidence", ""},
			},
			{
				Operation: "CREATE",
				Fields:    []string{"priority", "status"},
				Edges:     []string{"control"},
			},
			{
				Operation: "UPDATE",
				Fields:    []string{"name"},
			},
		},
	}

	operations, fields := DeriveTriggerPrefilter(doc)

	assert.Check(t, is.DeepEqual([]string{"CREATE", "UPDATE"}, operations))
	assert.Check(t, is.DeepEqual([]string{"control", "evidence", "name", "priority", "status"}, fields))
}

// TestDeriveTriggerPrefilterAnyFieldTrigger verifies any-field triggers clear fields
func TestDeriveTriggerPrefilterAnyFieldTrigger(t *testing.T) {
	doc := models.WorkflowDefinitionDocument{
		Triggers: []models.WorkflowTrigger{
			{
				Operation: "UPDATE",
				Fields:    []string{"status"},
			},
			{
				Operation: "UPDATE",
			},
			{
				Operation: "CREATE",
				Edges:     []string{""},
			},
		},
	}

	operations, fields := DeriveTriggerPrefilter(doc)

	assert.Check(t, is.DeepEqual([]string{"CREATE", "UPDATE"}, operations))
	assert.Check(t, is.Len(fields, 0))
}
