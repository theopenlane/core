package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// WorkflowApprovalMixin adds fields for storing proposed changes that require approval.
// This enables the "proposed changes" pattern where mutations are intercepted and stored
// for workflow approval before being applied to the actual entity.
type WorkflowApprovalMixin struct {
	mixin.Schema
}

// Fields of the ApprovalRequiredMixin
func (WorkflowApprovalMixin) Fields() []ent.Field {
	return []ent.Field{
		field.JSON("proposed_changes", map[string]any{}).
			Optional().
			Comment("pending changes awaiting workflow approval"),
		field.String("proposed_by_user_id").
			Optional().
			Comment("user who proposed the changes"),
		field.Time("proposed_at").
			Optional().
			Nillable().
			Comment("when changes were proposed"),
	}
}
