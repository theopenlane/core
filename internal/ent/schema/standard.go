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

// Standard defines the standard schema.
type Standard struct {
	ent.Schema
}

// Fields returns standard fields.
func (Standard) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name of the standard body, e.g. TSC, NIST, SOC, HITRUST, FedRamp, etc."),
		field.Text("description").
			Optional().
			Comment("description of the standard"),
		field.String("family").
			Optional().
			Comment("family of the standard, e.g. 800-53, 800-171, 27001, etc."),
		field.String("status").
			Optional().
			Comment("status of the standard - active, deprecated, etc."),
		field.String("type").
			Optional().
			Comment("type of the standard - security, privacy, etc."),
		field.String("version").
			Optional().
			Comment("version of the standard"),
		field.Text("purpose and scope").
			Optional().
			Comment("purpose and scope"),
		field.Text("background").
			Optional().
			Comment("background"),
		field.Text("satisfies").
			Optional().
			Comment("which controls are satisfied by the standard"),
		field.JSON("jsonschema", map[string]interface{}{}).
			Optional().
			Comment("json schema"),
	}
}

// Edges of the Standard
func (Standard) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("controlobjective", ControlObjective.Type),
		edge.To("control", Control.Type),
		edge.To("procedure", Procedure.Type),
		edge.To("actionplan", ActionPlan.Type),
	}
}

// Mixin of the Standard
func (Standard) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		mixin.SoftDeleteMixin{},
		emixin.IDMixin{},
		emixin.TagMixin{},
	}
}

// Annotations of the Standard
func (Standard) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
	}
}