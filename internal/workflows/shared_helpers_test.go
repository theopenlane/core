package workflows

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
)

func TestWorkflowCreationError(t *testing.T) {
	baseErr := errors.New("boom")
	err := &WorkflowCreationError{Stage: WorkflowCreationStageInstance, Err: baseErr}

	assert.True(t, strings.Contains(err.Error(), "workflow instance creation failed"))
	assert.ErrorIs(t, err, baseErr)
}

func TestBuildProposedChanges(t *testing.T) {
	m := fakeMutation{
		fields:  []string{"fieldA"},
		cleared: []string{"fieldB"},
		values: map[string]any{
			"fieldA": "value",
		},
	}

	changes := BuildProposedChanges(m, []string{"fieldA", "fieldB", "fieldC"})
	assert.Equal(t, map[string]any{
		"fieldA": "value",
		"fieldB": nil,
	}, changes)
}

func TestDefinitionMatchesTrigger(t *testing.T) {
	doc := models.WorkflowDefinitionDocument{
		Triggers: []models.WorkflowTrigger{
			{Operation: "UPDATE", Fields: []string{"name"}},
		},
	}
	assert.False(t, DefinitionMatchesTrigger(doc, "CREATE", []string{"name"}))
	assert.False(t, DefinitionMatchesTrigger(doc, "", []string{"name"}))

	doc = models.WorkflowDefinitionDocument{
		Triggers: []models.WorkflowTrigger{
			{Operation: "UPDATE"},
		},
	}
	assert.True(t, DefinitionMatchesTrigger(doc, "UPDATE", nil))

	doc = models.WorkflowDefinitionDocument{
		Triggers: []models.WorkflowTrigger{
			{Fields: []string{"status"}, Edges: []string{"owner"}},
		},
	}
	assert.True(t, DefinitionMatchesTrigger(doc, "", []string{"status"}))

	doc = models.WorkflowDefinitionDocument{
		Triggers: []models.WorkflowTrigger{
			{Fields: []string{"status"}, Edges: []string{"owner"}},
		},
	}
	assert.False(t, DefinitionMatchesTrigger(doc, "", []string{"other"}))
}

func TestResolveOwnerID(t *testing.T) {
	ownerID, err := ResolveOwnerID(context.Background(), "owner-1")
	assert.NoError(t, err)
	assert.Equal(t, "owner-1", ownerID)

	orgID := ulids.New().String()
	ctx := auth.NewTestContextWithOrgID(ulids.New().String(), orgID)
	ownerID, err = ResolveOwnerID(ctx, "")
	assert.NoError(t, err)
	assert.Equal(t, orgID, ownerID)

	_, err = ResolveOwnerID(context.Background(), "")
	assert.ErrorIs(t, err, auth.ErrNoAuthUser)
}

func TestBuildWorkflowActionContext(t *testing.T) {
	instance := &generated.WorkflowInstance{
		ID:                   "inst-1",
		WorkflowDefinitionID: "def-1",
	}
	obj := &Object{ID: "obj-1", Type: enums.WorkflowObjectTypeControl}

	replacements, data := BuildWorkflowActionContext(instance, obj, "action-1")
	assert.Equal(t, "inst-1", replacements["instance_id"])
	assert.Equal(t, "def-1", replacements["definition_id"])
	assert.Equal(t, "obj-1", replacements["object_id"])
	assert.Equal(t, obj.Type.String(), replacements["object_type"])
	assert.Equal(t, "action-1", replacements["action_key"])

	require.Contains(t, data, "instance_id")
	assert.Equal(t, replacements["object_id"], data["object_id"])
}

func TestValidateCELExpression(t *testing.T) {
	assert.NoError(t, ValidateCELExpression("1 + 1"))

	err := ValidateCELExpression("1 +")
	assert.Error(t, err)
}
