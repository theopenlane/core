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
			Comment("description of the procedure"),
		field.String("status").
			Comment("status of the procedure"),
		field.String("type").
			Comment("type of the procedure"),
		field.String("version").
			Comment("version of the procedure"),
		field.Text("purpose and scope").
			Comment("purpose and scope"),
		field.Text("background").
			Comment("background"),
		field.Text("satisfies").
			Comment("which controls are satisfied by the procedure"),
	}
}

// Edges of the Procedure
func (Procedure) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("control", User.Type).
			Ref("controls"),
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
