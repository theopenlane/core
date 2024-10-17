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

// ActionPlan defines the actionplan schema.
type ActionPlan struct {
	ent.Schema
}

// Fields returns actionplan fields.
func (ActionPlan) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name of the action plan"),
		field.Text("description").
			Comment("description of the action plan"),
		field.String("status").
			Comment("status of the action plan"),
		field.String("assigned").
			Comment("assigned to"),
		field.String("due_date").
			Comment("due date"),
		field.String("priority").
			Comment("priority"),
		field.String("source").
			Comment("source of the action plan"),
	}
}

// Edges of the ActionPlan
func (ActionPlan) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("standard", User.Type).
			Ref("actionplan"),
	}
}

// Mixin of the ActionPlan
func (ActionPlan) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		mixin.SoftDeleteMixin{},
		emixin.IDMixin{},
		emixin.TagMixin{},
	}
}

// Annotations of the ActionPlan
func (ActionPlan) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
	}
}
