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
			Comment("the name of the control objective"),
		field.Text("description").
			Optional().
			Comment("description of the control objective"),
		field.String("status").
			Optional().
			Comment("status of the control objective"),
		field.String("control_objective_type").
			Optional().
			Comment("type of the control objective"),
		field.String("version").
			Optional().
			Comment("version of the control objective"),
		field.String("control_number").
			Optional().
			Comment("number of the control objective"),
		field.Text("family").
			Optional().
			Comment("family of the control objective"),
		field.String("class").
			Optional().
			Comment("class associated with the control objective"),
		field.String("source").
			Optional().
			Comment("source of the control objective, e.g. framework, template, user-defined, etc."),
		field.Text("mapped_frameworks").
			Optional().
			Comment("mapped frameworks"),
		field.JSON("details", map[string]interface{}{}).
			Optional().
			Comment("json data including details of the control objective"),
	}
}

// Edges of the ControlObjective
func (ControlObjective) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("policy", InternalPolicy.Type).
			Ref("controlobjectives"),
		edge.To("controls", Control.Type),
		edge.To("procedures", Procedure.Type),
		edge.To("risks", Risk.Type),
		edge.To("subcontrols", Subcontrol.Type),
		edge.From("standard", Standard.Type).
			Ref("controlobjectives"),
		edge.To("narratives", Narrative.Type),
		edge.To("tasks", Task.Type),
		edge.From("programs", Program.Type).
			Ref("controlobjectives"),
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
