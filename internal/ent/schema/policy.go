package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	emixin "github.com/theopenlane/entx/mixin"

	"github.com/theopenlane/core/internal/ent/mixin"
)

// Policy defines the policy schema.
type Policy struct {
	ent.Schema
}

// Fields returns policy fields.
func (Policy) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name of the policy"),
		field.Text("description").
			Comment("description of the policy"),
		field.String("status").
			Comment("status of the policy"),
		field.String("type").
			Comment("type of the policy"),
		field.String("version").
			Comment("version of the policy"),
		field.Text("purpose and scope").
			Comment("purpose and scope"),
		field.Text("background").
			Comment("background"),
	}
}

// Edges of the Policy
func (Policy) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("controlobjective", ControlObjective.Type),
	}
}

// Mixin of the Policy
func (Policy) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		mixin.SoftDeleteMixin{},
		emixin.IDMixin{},
		emixin.TagMixin{},
	}
}

// Annotations of the Policy
func (Policy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
	}
}
