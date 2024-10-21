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

// InternalPolicy defines the policy schema.
type InternalPolicy struct {
	ent.Schema
}

// Fields returns policy fields.
func (InternalPolicy) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name of the policy"),
		field.Text("description").
			Comment("description of the policy"),
		field.String("status").
			Optional().
			Comment("status of the policy"),
		field.String("policy_type").
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

// Edges of the InternalPolicy
func (InternalPolicy) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("controlobjectives", ControlObjective.Type),
		edge.To("controls", Control.Type),
		edge.To("procedures", Procedure.Type),
		edge.To("narratives", Narrative.Type),
	}
}

// Mixin of the InternalPolicy
func (InternalPolicy) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		mixin.SoftDeleteMixin{},
		emixin.IDMixin{},
		emixin.TagMixin{},
	}
}

// Annotations of the InternalPolicy
func (InternalPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
	}
}
