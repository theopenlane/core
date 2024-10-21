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

// Narrative defines the narrative schema.
type Narrative struct {
	ent.Schema
}

// Fields returns narrative fields.
func (Narrative) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name of the file provided in the payload key without the extension"),
		field.Text("description").
			Optional().
			Comment("the description of the narrative"),
		field.Text("satisfies").
			Optional().
			Comment("which controls are satisfied by the narrative"),
		field.JSON("jsonschema", map[string]interface{}{}).
			Optional().
			Comment("json schema"),
	}
}

// Edges of the Narrative
func (Narrative) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("policy", Policy.Type),
		edge.From("control", Control.Type),
		edge.From("procedure", Procedure.Type),
		edge.From("controlobjective", ControlObjective.Type),
	}
}

// Mixin of the Narrative
func (Narrative) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		mixin.SoftDeleteMixin{},
		emixin.IDMixin{},
		emixin.TagMixin{},
	}
}

// Annotations of the Narrative
func (Narrative) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
	}
}
