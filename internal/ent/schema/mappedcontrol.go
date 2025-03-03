package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/theopenlane/core/internal/ent/mixin"
	emixin "github.com/theopenlane/entx/mixin"
)

// MappedControl holds the schema definition for the MappedControl entity
type MappedControl struct {
	ent.Schema
}

// Fields of the MappedControls
func (MappedControl) Fields() []ent.Field {
	return []ent.Field{
		field.String("control_id").
			Comment("the id of the control being mapped").
			Unique().
			Immutable().
			NotEmpty(),
		field.String("mapped_control_id").
			Unique().
			Immutable().
			NotEmpty().
			Comment("the id of the control that is mapped to"),
		field.String("mapping_type").
			Comment("the type of mapping between the two controls, e.g. subset, intersect, equal, superset").
			Optional(),
		field.String("relation").
			Comment("description of how the two controls are related").
			Optional(),
	}
}

// Mixin of the MappedControls
func (MappedControl) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		emixin.IDMixin{},
		mixin.SoftDeleteMixin{},
		emixin.TagMixin{},
	}
}

// Edges of the MappedControls
func (MappedControl) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("control", Control.Type).
			Unique().
			Required().
			Field("control_id").
			Immutable(),
		edge.To("mapped_control", Control.Type).
			Comment("mapped control to the original control, meaning there is overlap between the controls").
			Unique().
			Required().
			Field("mapped_control_id").
			Immutable(),
	}
}

// Indexes of the MappedControls
func (MappedControl) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("control_id", "mapped_control_id").
			Unique(),
	}
}

// Annotations of the MappedControls
func (MappedControl) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entgql.RelayConnection(),
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		// entfga.MembershipChecks("control"),
	}
}

// Hooks of the MappedControls
func (MappedControl) Hooks() []ent.Hook {
	return []ent.Hook{}
}

// Interceptors of the MappedControls
func (MappedControl) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}

// // Policy of the MappedControls
// func (MappedControls) Policy() ent.Policy {
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
// 			entfga.CheckEditAccess[*generated.MappedControlsMutation](),
// 		),
// 	)
// }
