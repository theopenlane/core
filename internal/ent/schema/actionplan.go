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
			Optional().
			Comment("description of the action plan"),
		field.String("status").
			Optional().
			Comment("status of the action plan"),
		field.Time("due_date").
			Optional().
			Comment("due date of the action plan"),
		field.String("priority").
			Optional().
			Comment("priority of the action plan"),
		field.String("source").
			Optional().
			Comment("source of the action plan"),
		field.JSON("details", map[string]interface{}{}).
			Optional().
			Comment("json data including details of the action plan"),
	}
}

// Edges of the ActionPlan
func (ActionPlan) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("standard", Standard.Type).
			Ref("actionplans"),
		edge.From("risk", Risk.Type).
			Ref("actionplans"),
		edge.From("control", Control.Type).
			Ref("actionplans"),
		edge.From("user", User.Type).
			Ref("actionplans"),
		edge.From("program", Program.Type).
			Ref("actionplans"),
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
