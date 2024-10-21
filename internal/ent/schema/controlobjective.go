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

// ControlObjective defines the controlobjective schema.
type ControlObjective struct {
	ent.Schema
}

// Fields returns controlobjective fields.
func (ControlObjective) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name of the control"),
		field.Text("description").
			Optional().
			Comment("description of the control"),
		field.String("status").
			Optional().
			Comment("status of the control"),
		field.String("type").
			Optional().
			Comment("type of the control"),
		field.String("version").
			Optional().
			Comment("version of the control"),
		field.String("owner").
			Optional().
			Comment("owner of the control"),
		field.String("control_number").
			Optional().
			Comment("control number"),
		field.Text("control_family").
			Optional().
			Comment("control family"),
		field.String("control_class").
			Optional().
			Comment("control class"),
		field.String("source").
			Optional().
			Comment("source of the control"),
		field.Text("mapped_frameworks").
			Optional().
			Comment("mapped frameworks"),
		field.JSON("jsonschema", map[string]interface{}{}).
			Optional().
			Comment("json schema"),
	}
}

// Edges of the ControlObjective
func (ControlObjective) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("policy", Policy.Type).
			Ref("policies"),
		edge.To("control", Control.Type),
		edge.To("procedure", Procedure.Type),
		edge.To("risk", Risk.Type),
		edge.To("subcontrol", Subcontrol.Type),
		edge.From("standard", Standard.Type).
			Ref("standards"),
	}
}

// Mixin of the ControlObjective
func (ControlObjective) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		mixin.SoftDeleteMixin{},
		emixin.IDMixin{},
		emixin.TagMixin{},
	}
}

// Annotations of the ControlObjective
func (ControlObjective) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
	}
}
