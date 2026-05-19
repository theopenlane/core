package rule

import (
	"testing"

	"entgo.io/ent"
	"github.com/stretchr/testify/assert"
)

func TestScopedRelation(t *testing.T) {
	tests := []struct {
		name       string
		objectType string
		relation   string
		op         ent.Op
		expected   string
	}{
		{name: "view query", objectType: "Control", relation: "can_view", expected: "can_view_control"},
		{name: "update op", objectType: "Task", relation: "can_view", op: ent.OpUpdate, expected: "can_edit_task"},
		{name: "edit relation", objectType: "Evidence", relation: "can_edit", expected: "can_edit_evidence"},
		{name: "delete op", objectType: "File", relation: "can_edit", op: ent.OpDeleteOne, expected: "can_delete_file"},
		{name: "create op", objectType: "Program", op: ent.OpCreate, expected: "can_create_program"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			actual := scopedRelation(tt.objectType, tt.relation, &tt.op)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
