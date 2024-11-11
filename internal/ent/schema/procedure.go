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

// Procedure defines the procedure schema.
type Procedure struct {
	ent.Schema
}

// Fields returns procedure fields.
func (Procedure) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name of the procedure"),
		field.Text("description").
			Optional().
			Comment("description of the procedure"),
		field.String("status").
			Optional().
			Comment("status of the procedure"),
		field.String("procedure_type").
			Optional().
			Comment("type of the procedure"),
		field.String("version").
			Optional().
			Comment("version of the procedure"),
		field.Text("purpose_and_scope").
			Optional().
			Comment("purpose and scope"),
		field.Text("background").
			Optional().
			Comment("background of the procedure"),
		field.Text("satisfies").
			Optional().
			Comment("which controls are satisfied by the procedure"),
		field.JSON("details", map[string]interface{}{}).
			Optional().
			Comment("json data for the procedure document"),
	}
}

// Edges of the Procedure
func (Procedure) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("control", Control.Type).
			Ref("procedures"),
		edge.From("internalpolicy", InternalPolicy.Type).
			Ref("procedures"),
		edge.To("narratives", Narrative.Type),
		edge.To("risks", Risk.Type),
		edge.To("tasks", Task.Type),
		edge.From("programs", Program.Type).
			Ref("procedures"),
	}
}

// Mixin of the Procedure
func (Procedure) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		mixin.SoftDeleteMixin{},
		emixin.IDMixin{},
		emixin.TagMixin{},
	}
}

// Annotations of the Procedure
func (Procedure) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
	}
}
