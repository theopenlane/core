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
		field.String("assigned").
			Optional().
			Comment("assigned to"),
		field.String("due_date").
			Optional().
			Comment("due date"),
		field.String("priority").
			Optional().
			Comment("priority"),
		field.String("source").
			Optional().
			Comment("source of the action plan"),
		field.JSON("jsonschema", map[string]interface{}{}).
			Optional().
			Comment("json schema"),
	}
}

// Edges of the ActionPlan
func (ActionPlan) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("standard", User.Type).
			Ref("standards"),
		edge.From("risk", Risk.Type).
			Ref("risks"),
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
