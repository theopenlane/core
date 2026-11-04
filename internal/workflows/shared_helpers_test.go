package workflows

import (
	"errors"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
)

func TestWorkflowCreationError(t *testing.T) {
	baseErr := errors.New("boom")
	err := &WorkflowCreationError{Stage: WorkflowCreationStageInstance, Err: baseErr}

	assert.Check(t, strings.Contains(err.Error(), "workflow instance creation failed"))
	assert.Check(t, is.ErrorIs(err, baseErr))
}

func TestBuildWorkflowActionContext(t *testing.T) {
	instance := &generated.WorkflowInstance{
		ID:                   "inst-1",
		WorkflowDefinitionID: "def-1",
	}
	obj := &Object{ID: "obj-1", Type: enums.WorkflowObjectTypeControl}

	replacements, data := BuildWorkflowActionContext(instance, obj, "action-1")
	assert.Check(t, is.Equal("inst-1", replacements["instance_id"]))
	assert.Check(t, is.Equal("def-1", replacements["definition_id"]))
	assert.Check(t, is.Equal("obj-1", replacements["object_id"]))
	assert.Check(t, is.Equal(obj.Type.String(), replacements["object_type"]))
	assert.Check(t, is.Equal("action-1", replacements["action_key"]))

	assert.Check(t, is.Contains(data, "instance_id"))
	assert.Check(t, is.Equal(replacements["object_id"], data["object_id"]))
}

func TestValidateCELExpression(t *testing.T) {
	assert.Check(t, ValidateCELExpression(nil, CELScopeAction, "1 + 1"))

	err := ValidateCELExpression(nil, CELScopeAction, "1 +")
	assert.Check(t, err != nil)
}
