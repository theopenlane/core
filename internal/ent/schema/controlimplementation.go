package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"

	"github.com/theopenlane/core/internal/ent/mixin"
	emixin "github.com/theopenlane/entx/mixin"
)

// ControlImplementation holds the schema definition for the ControlImplementation entity
type ControlImplementation struct {
	ent.Schema
}

// Fields of the ControlImplementation
func (ControlImplementation) Fields() []ent.Field {
	return []ent.Field{
		field.String("control_id").
			Comment("the id of the control that this implementation is for").
			NotEmpty(),
		field.String("status").
			Optional().
			Comment("status of the control implementation"),
		field.Time("implementation_date").
			Optional().
			Comment("date the control was implemented"),
		field.Bool("verified").
			Optional().
			Comment("set to true if the control implementation has been verified"),
		field.Time("verification_date").
			Optional().
			Comment("date the control implementation was verified"),
		field.Text("details").
			Optional().
			Comment("details of the control implementation"),
	}
}

// Mixin of the ControlImplementation
func (ControlImplementation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		emixin.IDMixin{},
		mixin.SoftDeleteMixin{},
		emixin.TagMixin{},
	}
}

// Edges of the ControlImplementation
func (ControlImplementation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("control", Control.Type).
			Ref("implementation"),
	}
}

// Indexes of the ControlImplementation
func (ControlImplementation) Indexes() []ent.Index {
	return []ent.Index{}
}

// Annotations of the ControlImplementation
func (ControlImplementation) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entgql.RelayConnection(),
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		// entfga.MembershipChecks("control"),
	}
}

// Hooks of the ControlImplementation
func (ControlImplementation) Hooks() []ent.Hook {
	return []ent.Hook{}
}

// Interceptors of the ControlImplementation
func (ControlImplementation) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}

// // Policy of the ControlImplementation
// func (ControlImplementation) Policy() ent.Policy {
// 	// add the new policy here, the default post-policy is to deny all
// 	// so you need to ensure there are rules in place to allow the actions you want
// 	return policy.NewPolicy(
// 		policy.WithQueryRules(
// 			// add query rules here, the below is the recommended default
// 			privacy.AlwaysAllowRule(), //  interceptor should filter out the results
// 		),
// 		policy.WithMutationRules(
// 			// add mutation rules here, the below is the recommended default
// 			policy.CheckCreateAccess(),
// 			entfga.CheckEditAccess[*generated.ControlImplementationMutation](),
// 		),
// 	)
// }
