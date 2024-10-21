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
			Optional().
			Comment("description of the subcontrol"),
		field.String("status").
			Optional().
			Comment("status of the subcontrol"),
		field.String("type").
			Optional().
			Comment("type of the subcontrol"),
		field.String("version").
			Optional().
			Comment("version of the control"),
		field.String("owner").
			Optional().
			Comment("owner of the subcontrol"),
		field.String("subcontrol_number").
			Optional().
			Comment("control number"),
		field.Text("subcontrol_family").
			Optional().
			Comment("control family"),
		field.String("subcontrol_class").
			Optional().
			Comment("control class"),
		field.String("source").
			Optional().
			Comment("source of the control"),
		field.Text("mapped_frameworks").
			Optional().
			Comment("mapped frameworks"),
		field.String("assigned_to").
			Optional().
			Comment("assigned to"),
		field.String("implementation_status").
			Optional().
			Comment("implementation status"),
		field.String("implementation_notes").
			Optional().
			Comment("implementation notes"),
		field.String("implementation_date").
			Optional().
			Comment("implementation date"),
		field.String("implementation_evidence").
			Optional().
			Comment("implementation evidence"),
		field.String("implementation_verification").
			Optional().
			Comment("implementation verification"),
		field.String("implementation_verification_date").
			Optional().
			Comment("implementation verification date"),
		field.JSON("jsonschema", map[string]interface{}{}).
			Optional().
			Comment("json schema"),
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
