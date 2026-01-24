package workflows

import (
	"testing"

	"github.com/stretchr/testify/assert"

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

	assert.Equal(t, []string{"CREATE", "UPDATE"}, operations)
	assert.Equal(t, []string{"control", "evidence", "name", "priority", "status"}, fields)
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

	assert.Equal(t, []string{"CREATE", "UPDATE"}, operations)
	assert.Empty(t, fields)
}
