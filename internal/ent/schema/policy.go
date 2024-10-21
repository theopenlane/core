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
			Optional().
			Comment("status of the policy"),
		field.String("type").
			Optional().
			Comment("type of the policy"),
		field.String("version").
			Optional().
			Comment("version of the policy"),
		field.Text("purpose and scope").
			Optional().
			Comment("purpose and scope"),
		field.Text("background").
			Optional().
			Comment("background"),
		field.JSON("jsonschema", map[string]interface{}{}).
			Optional().
			Comment("json schema"),
	}
}

// Edges of the Policy
func (Policy) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("controlobjective", ControlObjective.Type),
		edge.To("control", Control.Type),
		edge.To("procedure", Procedure.Type),
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
