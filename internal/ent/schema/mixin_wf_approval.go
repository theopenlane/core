package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/theopenlane/entx"
)

// WorkflowApprovalMixin adds fields for storing proposed changes that require approval.
// This enables the "proposed changes" pattern where mutations are intercepted and stored
// for workflow approval before being applied to the actual entity.
type WorkflowApprovalMixin struct {
	mixin.Schema
}

// Fields of the WorkflowApprovalMixin
func (WorkflowApprovalMixin) Fields() []ent.Field {
	return []ent.Field{
		// Marker field to indicate this schema is workflow-eligible
		// This field is used by templates to detect schemas with WorkflowApprovalMixin
		field.Bool("workflow_eligible_marker").
			Annotations(entx.FieldWorkflowEligible()).
			Optional().
			Default(true).
			StructTag(`json:"-"`).
			Comment("internal marker field for workflow eligibility, not exposed in API"),
	}
}

// Interceptors of the WorkflowApprovalMixin
// func (WorkflowApprovalMixin) Interceptors() []ent.Interceptor {
// 	return []ent.Interceptor{
// 		interceptors.WorkflowApprovalInterceptor(),
// 	}
// }
