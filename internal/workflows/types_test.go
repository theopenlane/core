package workflows

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/workflowobjectref"
	"github.com/theopenlane/core/v2/internal/workflows/observability"
)

type workflowTestNode struct{}

func (*workflowTestNode) IsNode() {}

// TestCELValue verifies CEL value construction for objects
func TestCELValue(t *testing.T) {
	var nilObj *Object
	assert.Check(t, is.Nil(nilObj.CELValue()))

	obj := &Object{ID: "obj1", Type: enums.WorkflowObjectTypeControl}
	value := obj.CELValue()
	_, isMap := value.(map[string]any)
	assert.Check(t, isMap)

	mapped := value.(map[string]any)
	assert.Check(t, is.Equal("obj1", mapped["id"]))
	assert.Check(t, is.Equal(enums.WorkflowObjectTypeControl, mapped["type"]))

	node := &workflowTestNode{}
	obj.Node = node
	assert.Check(t, is.Equal(node, obj.CELValue()))
}

// TestBuildCELVarsDefault verifies default CEL variable building
func TestBuildCELVarsDefault(t *testing.T) {
	old := celContextBuilders
	t.Cleanup(func() { celContextBuilders = old })
	celContextBuilders = nil

	obj := &Object{ID: "obj1", Type: enums.WorkflowObjectTypeControl}
	proposed := map[string]any{"status": "APPROVED"}
	vars := BuildCELVars(obj, []string{"status"}, []string{"edge"}, map[string][]string{"a": []string{"1"}}, map[string][]string{"b": []string{"2"}}, "UPDATE", "user1", proposed)

	assert.Check(t, is.DeepEqual(obj.CELValue(), vars["object"]))
	assert.Check(t, is.DeepEqual([]string{"status"}, vars["changed_fields"]))
	assert.Check(t, is.DeepEqual([]string{"edge"}, vars["changed_edges"]))
	assert.Check(t, is.DeepEqual(map[string][]string{"a": []string{"1"}}, vars["added_ids"]))
	assert.Check(t, is.DeepEqual(map[string][]string{"b": []string{"2"}}, vars["removed_ids"]))
	assert.Check(t, is.Equal("UPDATE", vars["event_type"]))
	assert.Check(t, is.Equal("user1", vars["user_id"]))
	assert.Check(t, is.DeepEqual(proposed, vars["proposed_changes"]))
}

// TestBuildCELVarsCustomBuilder verifies custom CEL context builders
func TestBuildCELVarsCustomBuilder(t *testing.T) {
	old := celContextBuilders
	t.Cleanup(func() { celContextBuilders = old })
	celContextBuilders = nil

	expected := map[string]any{"custom": true}
	RegisterCELContextBuilder(func(_ *Object, _ []string, _ []string, _ map[string][]string, _ map[string][]string, _ string, _ string, _ map[string]any) map[string]any {
		return expected
	})

	vars := BuildCELVars(nil, nil, nil, nil, nil, "", "", nil)
	assert.Check(t, is.DeepEqual(expected, vars))
}

// TestObjectFromRef verifies object lookup from references
func TestObjectFromRef(t *testing.T) {
	_, err := ObjectFromRef(&generated.WorkflowObjectRef{})
	assert.Check(t, is.ErrorIs(err, ErrMissingObjectID))

	obj, err := ObjectFromRef(&generated.WorkflowObjectRef{ControlID: "obj1"})
	assert.Check(t, err)
	assert.Check(t, is.DeepEqual(&Object{ID: "obj1", Type: enums.WorkflowObjectTypeControl}, obj))
}

// TestBuildAssignmentContext verifies assignment context construction
func TestBuildAssignmentContext(t *testing.T) {
	old := assignmentContextBuilder
	t.Cleanup(func() { assignmentContextBuilder = old })
	assignmentContextBuilder = nil

	ctx := context.Background()
	vars, err := BuildAssignmentContext(ctx, nil, "")
	assert.Check(t, err)
	assert.Check(t, is.Nil(vars))

	assignmentContextBuilder = func(_ context.Context, _ *generated.Client, instanceID string) (map[string]any, error) {
		return map[string]any{"instance_id": instanceID}, nil
	}

	vars, err = BuildAssignmentContext(ctx, nil, "instance-1")
	assert.Check(t, err)
	assert.Check(t, is.DeepEqual(map[string]any{"instance_id": "instance-1"}, vars))
}

func TestRegisterAssignmentContextBuilder(t *testing.T) {
	old := assignmentContextBuilder
	t.Cleanup(func() { assignmentContextBuilder = old })

	RegisterAssignmentContextBuilder(func(_ context.Context, _ *generated.Client, instanceID string) (map[string]any, error) {
		return map[string]any{"instance_id": instanceID}, nil
	})

	out, err := BuildAssignmentContext(context.Background(), nil, "instance-2")
	assert.Check(t, err)
	assert.Check(t, is.DeepEqual(map[string]any{"instance_id": "instance-2"}, out))
}

func TestObservabilityFields(t *testing.T) {
	var nilObj *Object
	assert.Check(t, is.Nil(nilObj.ObservabilityFields()))

	// Register an observability fields builder that matches the expected behavior
	old := observabilityFieldsBuilder
	RegisterObservabilityFieldsBuilder(func(obj *Object) map[string]any {
		if obj == nil {
			return nil
		}
		fields := map[string]any{
			observability.FieldObjectType: obj.Type.String(),
		}
		switch obj.Type {
		case enums.WorkflowObjectTypeActionPlan:
			fields[workflowobjectref.FieldActionPlanID] = obj.ID
		case enums.WorkflowObjectTypeControl:
			fields[workflowobjectref.FieldControlID] = obj.ID
		case enums.WorkflowObjectTypeEvidence:
			fields[workflowobjectref.FieldEvidenceID] = obj.ID
		case enums.WorkflowObjectTypeInternalPolicy:
			fields[workflowobjectref.FieldInternalPolicyID] = obj.ID
		case enums.WorkflowObjectTypeProcedure:
			fields[workflowobjectref.FieldProcedureID] = obj.ID
		case enums.WorkflowObjectTypeSubcontrol:
			fields[workflowobjectref.FieldSubcontrolID] = obj.ID
		}
		return fields
	})
	t.Cleanup(func() { observabilityFieldsBuilder = old })

	cases := map[enums.WorkflowObjectType]string{
		enums.WorkflowObjectTypeActionPlan:     workflowobjectref.FieldActionPlanID,
		enums.WorkflowObjectTypeControl:        workflowobjectref.FieldControlID,
		enums.WorkflowObjectTypeEvidence:       workflowobjectref.FieldEvidenceID,
		enums.WorkflowObjectTypeInternalPolicy: workflowobjectref.FieldInternalPolicyID,
		enums.WorkflowObjectTypeProcedure:      workflowobjectref.FieldProcedureID,
		enums.WorkflowObjectTypeSubcontrol:     workflowobjectref.FieldSubcontrolID,
	}

	for objectType, fieldName := range cases {
		obj := &Object{ID: "obj", Type: objectType}
		fields := obj.ObservabilityFields()
		assert.Check(t, is.Equal(objectType.String(), fields[observability.FieldObjectType]))
		assert.Check(t, is.Equal("obj", fields[fieldName]))
	}

	obj := &Object{ID: "obj", Type: enums.WorkflowObjectTypeCampaign}
	fields := obj.ObservabilityFields()
	assert.Check(t, is.Equal(obj.Type.String(), fields[observability.FieldObjectType]))
	assert.Check(t, is.Len(fields, 1))
}
