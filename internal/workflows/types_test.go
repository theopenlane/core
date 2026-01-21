package workflows

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
)

func TestCELValue(t *testing.T) {
	var nilObj *Object
	assert.Nil(t, nilObj.CELValue())

	obj := &Object{ID: "obj1", Type: enums.WorkflowObjectTypeControl}
	value := obj.CELValue()
	require.IsType(t, map[string]any{}, value)

	mapped := value.(map[string]any)
	assert.Equal(t, "obj1", mapped["id"])
	assert.Equal(t, enums.WorkflowObjectTypeControl, mapped["type"])

	node := map[string]any{"k": "v"}
	obj.Node = node
	assert.Equal(t, node, obj.CELValue())
}

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
	require.NoError(t, err)
	assert.Equal(t, expected, obj)
}

func TestBuildAssignmentContext(t *testing.T) {
	old := assignmentContextBuilder
	t.Cleanup(func() { assignmentContextBuilder = old })
	assignmentContextBuilder = nil

	ctx := context.Background()
	vars, err := BuildAssignmentContext(ctx, nil, "")
	require.NoError(t, err)
	assert.Nil(t, vars)

	assignmentContextBuilder = func(_ context.Context, _ *generated.Client, instanceID string) (map[string]any, error) {
		return map[string]any{"instance_id": instanceID}, nil
	}

	vars, err = BuildAssignmentContext(ctx, nil, "instance-1")
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"instance_id": "instance-1"}, vars)
}
