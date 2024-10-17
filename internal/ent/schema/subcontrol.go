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

// Subcontrol defines the file schema.
type Subcontrol struct {
	ent.Schema
}

// Fields returns file fields.
func (Subcontrol) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name of the subcontrol"),
		field.Text("description").
			Comment("description of the subcontrol"),
		field.String("status").
			Comment("status of the subcontrol"),
		field.String("type").
			Comment("type of the subcontrol"),
		field.String("version").
			Comment("version of the control"),
		field.String("owner").
			Comment("owner of the subcontrol"),
		field.String("subcontrol_number").
			Comment("control number"),
		field.Text("subcontrol_family").
			Comment("control family"),
		field.String("subcontrol_class").
			Comment("control class"),
		field.String("source").
			Comment("source of the control"),
		field.Text("mapped_frameworks").
			Comment("mapped frameworks"),
		field.String("assigned_to").
			Comment("assigned to"),
		field.String("implementation_status").
			Comment("implementation status"),
		field.String("implementation_notes").
			Comment("implementation notes"),
		field.String("implementation_date").
			Comment("implementation date"),
		field.String("implementation_evidence").
			Comment("implementation evidence"),
		field.String("implementation_verification").
			Comment("implementation verification"),
		field.String("implementation_verification_date").
			Comment("implementation verification date"),
	}
}

// Edges of the Subcontrol
func (Subcontrol) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("control", Control.Type).
			Ref("controls"),
	}
}

// Mixin of the Subcontrol
func (Subcontrol) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		mixin.SoftDeleteMixin{},
		emixin.IDMixin{},
		emixin.TagMixin{},
	}
}

// Annotations of the Subcontrol
func (Subcontrol) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
	}
}
