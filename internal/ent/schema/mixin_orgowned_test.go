package schema

import (
	"context"
	"testing"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
)

func TestSkipUserParentTupleFunc(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		op   ent.Op
	}{
		{
			name: "create",
			op:   ent.OpCreate,
		},
		{
			name: "update one",
			op:   ent.OpUpdateOne,
		},
		{
			name: "update many",
			op:   ent.OpUpdate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &generated.IntegrationRunMutation{}
			m.SetOp(tt.op)

			if !skipUserParentTupleFunc(context.Background(), m) {
				t.Fatalf("skipUserParentTupleFunc(%s) = false, want true", tt.op)
			}
		})
	}
}
