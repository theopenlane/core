package workflows

import (
	"context"
	"testing"

	"entgo.io/ent"
	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/workflowobjectref"
	"github.com/theopenlane/core/internal/workflows/observability"
)

// TestCELValue verifies CEL value construction for objects
func TestCELValue(t *testing.T) {
	var nilObj *Object
	assert.Nil(t, nilObj.CELValue())

	obj := &Object{ID: "obj1", Type: enums.WorkflowObjectTypeControl}
	value := obj.CELValue()
	assert.IsType(t, map[string]any{}, value)

	mapped := value.(map[string]any)
	assert.Equal(t, "obj1", mapped["id"])
	assert.Equal(t, enums.WorkflowObjectTypeControl, mapped["type"])

	node := map[string]any{"k": "v"}
	obj.Node = node
	assert.Equal(t, node, obj.CELValue())
}

// TestBuildCELVarsDefault verifies default CEL variable building
func TestBuildCELVarsDefault(t *testing.T) {
	old := celContextBuilders
	t.Cleanup(func() { celContextBuilders = old })
	celContextBuilders = nil

	obj := &Object{ID: "obj1", Type: enums.WorkflowObjectTypeControl}
	vars := BuildCELVars(obj, []string{"status"}, []string{"edge"}, map[string][]string{"a": []string{"1"}}, map[string][]string{"b": []string{"2"}}, "UPDATE", "user1")

	assert.Equal(t, obj.CELValue(), vars["object"])
	assert.Equal(t, []string{"status"}, vars["changed_fields"])
	assert.Equal(t, []string{"edge"}, vars["changed_edges"])
	assert.Equal(t, map[string][]string{"a": []string{"1"}}, vars["added_ids"])
	assert.Equal(t, map[string][]string{"b": []string{"2"}}, vars["removed_ids"])
	assert.Equal(t, "UPDATE", vars["event_type"])
	assert.Equal(t, "user1", vars["user_id"])
}

// TestBuildCELVarsCustomBuilder verifies custom CEL context builders
func TestBuildCELVarsCustomBuilder(t *testing.T) {
	old := celContextBuilders
	t.Cleanup(func() { celContextBuilders = old })
	celContextBuilders = nil

	expected := map[string]any{"custom": true}
	RegisterCELContextBuilder(func(_ *Object, _ []string, _ []string, _ map[string][]string, _ map[string][]string, _ string, _ string) map[string]any {
		return expected
	})

	vars := BuildCELVars(nil, nil, nil, nil, nil, "", "")
	assert.Equal(t, expected, vars)
}

// TestObjectFromRef verifies object lookup from references
func TestObjectFromRef(t *testing.T) {
	old := objectFromRefRegistry
	t.Cleanup(func() { objectFromRefRegistry = old })
	objectFromRefRegistry = nil

	_, err := ObjectFromRef(&generated.WorkflowObjectRef{})
	assert.ErrorIs(t, err, ErrMissingObjectID)

	expected := &Object{ID: "obj1", Type: enums.WorkflowObjectTypeControl}
	objectFromRefRegistry = []func(*generated.WorkflowObjectRef) (*Object, bool){
		func(_ *generated.WorkflowObjectRef) (*Object, bool) {
			return expected, true
		},
	}

	obj, err := ObjectFromRef(&generated.WorkflowObjectRef{})
	assert.NoError(t, err)
	assert.Equal(t, expected, obj)
}

// TestBuildAssignmentContext verifies assignment context construction
func TestBuildAssignmentContext(t *testing.T) {
	old := assignmentContextBuilder
	t.Cleanup(func() { assignmentContextBuilder = old })
	assignmentContextBuilder = nil

	ctx := context.Background()
	vars, err := BuildAssignmentContext(ctx, nil, "")
	assert.NoError(t, err)
	assert.Nil(t, vars)

	assignmentContextBuilder = func(_ context.Context, _ *generated.Client, instanceID string) (map[string]any, error) {
		return map[string]any{"instance_id": instanceID}, nil
	}

	vars, err = BuildAssignmentContext(ctx, nil, "instance-1")
	assert.NoError(t, err)
	assert.Equal(t, map[string]any{"instance_id": "instance-1"}, vars)
}

func TestRegisterObjectRefResolver(t *testing.T) {
	old := objectFromRefRegistry
	t.Cleanup(func() { objectFromRefRegistry = old })
	objectFromRefRegistry = nil

	RegisterObjectRefResolver(func(*generated.WorkflowObjectRef) (*Object, bool) {
		return &Object{ID: "obj", Type: enums.WorkflowObjectTypeControl}, true
	})

	obj, err := ObjectFromRef(&generated.WorkflowObjectRef{})
	assert.NoError(t, err)
	assert.Equal(t, "obj", obj.ID)
}

func TestRegisterObjectRefQueryBuilder(t *testing.T) {
	old := objectRefQueryBuilders
	t.Cleanup(func() { objectRefQueryBuilders = old })
	objectRefQueryBuilders = nil

	RegisterObjectRefQueryBuilder(func(q *generated.WorkflowObjectRefQuery, _ *Object) (*generated.WorkflowObjectRefQuery, bool) {
		return q, true
	})

	query := &generated.WorkflowObjectRefQuery{}
	out := buildObjectRefQuery(query, &Object{ID: "obj"})
	assert.Equal(t, query, out)
}

func TestRegisterAssignmentContextBuilder(t *testing.T) {
	old := assignmentContextBuilder
	t.Cleanup(func() { assignmentContextBuilder = old })

	RegisterAssignmentContextBuilder(func(_ context.Context, _ *generated.Client, instanceID string) (map[string]any, error) {
		return map[string]any{"instance_id": instanceID}, nil
	})

	out, err := BuildAssignmentContext(context.Background(), nil, "instance-2")
	assert.NoError(t, err)
	assert.Equal(t, map[string]any{"instance_id": "instance-2"}, out)
}

func TestObservabilityFields(t *testing.T) {
	var nilObj *Object
	assert.Nil(t, nilObj.ObservabilityFields())

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
		assert.Equal(t, objectType.String(), fields[observability.FieldObjectType])
		assert.Equal(t, "obj", fields[fieldName])
	}

	obj := &Object{ID: "obj", Type: enums.WorkflowObjectTypeCampaign}
	fields := obj.ObservabilityFields()
	assert.Equal(t, obj.Type.String(), fields[observability.FieldObjectType])
	assert.Len(t, fields, 1)
}

type fakeMutation struct {
	typ     string
	fields  []string
	cleared []string
	values  map[string]any
}

func (f fakeMutation) ID() (string, bool) {
	return "", false
}

func (f fakeMutation) IDs(context.Context) ([]string, error) {
	return nil, nil
}

func (f fakeMutation) Type() string {
	return f.typ
}

func (f fakeMutation) Op() ent.Op {
	return ent.Op(0)
}

func (f fakeMutation) Client() *generated.Client {
	return nil
}

func (f fakeMutation) Field(name string) (ent.Value, bool) {
	value, ok := f.values[name]
	return value, ok
}

func (f fakeMutation) Fields() []string {
	return f.fields
}

func (f fakeMutation) ClearedFields() []string {
	return f.cleared
}
