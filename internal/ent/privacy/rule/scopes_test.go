package rule

import (
	"errors"
	"testing"

	"entgo.io/ent"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
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
		{name: "update op", objectType: "Task", op: ent.OpUpdate, expected: "can_edit_task"},
		{name: "edit relation", objectType: "Evidence", relation: "can_edit", expected: "can_edit_evidence"},
		{name: "delete op", objectType: "File", op: ent.OpDeleteOne, expected: "can_delete_file"},
		{name: "create op", objectType: "Program", op: ent.OpCreate, expected: "can_create_program"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := scopedRelation(tt.objectType, tt.relation, &tt.op)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

// TestCheckSubjectScopeOrgSupport asserts an org-scoped support session may read and write org data but
// may never delete; delete is denied at the scope rule rather than falling through to FGA
func TestCheckSubjectScopeOrgSupport(t *testing.T) {
	const orgID = "01JSPPRT000000000000000000"

	tests := []struct {
		name       string
		objectType string
		relation   string
		op         *ent.Op
		expected   error
	}{
		{name: "support can query", objectType: "Control", relation: "can_view", expected: privacy.Allow},
		{name: "support can create", objectType: "Control", op: lo.ToPtr(ent.OpCreate), expected: privacy.Allow},
		{name: "support can edit", objectType: "Control", op: lo.ToPtr(ent.OpUpdateOne), expected: privacy.Allow},
		{name: "support cannot delete", objectType: "Control", op: lo.ToPtr(ent.OpDeleteOne), expected: ErrRequiredScopeNotSet},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := auth.WithCaller(t.Context(), auth.NewOrgSupportCaller(orgID, "support-subject", "Support", "support@theopenlane.io"))

			err := CheckSubjectScope(ctx, tt.objectType, tt.relation, tt.op)
			assert.True(t, errors.Is(err, tt.expected), "expected %v, got %v", tt.expected, err)
		})
	}
}
