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

// Risk defines the risk schema.
type Risk struct {
	ent.Schema
}

// Fields returns risk fields.
func (Risk) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name of the risk"),
		field.Text("description").
			Optional().
			Comment("description of the risk"),
		field.String("status").
			Optional().
			Comment("status of the risk - mitigated or not, inflight, etc."),
		field.String("type").
			Optional().
			Comment("type of the risk, e.g. strategic, operational, financial, external, etc."),
		field.Text("business costs").
			Optional().
			Comment("business costs associated with the risk"),
		field.Text("impact").
			Optional().
			Comment("impact of the risk"),
		field.Text("likelihood").
			Optional().
			Comment("likelihood of the risk"),
		field.Text("mitigation").
			Optional().
			Comment("mitigation of the risk"),
		field.Text("satisfies").
			Optional().
			Comment("which controls are satisfied by the risk"),
		field.Text("severity").
			Optional().
			Comment("severity of the risk, e.g. high medium low"),
		field.JSON("jsonschema", map[string]interface{}{}).
			Optional().
			Comment("json schema"),
	}
}

// Edges of the Risk
func (Risk) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("control", Control.Type).
			Ref("controls"),
		edge.From("procedure", Procedure.Type).
			Ref("procedures"),
		edge.To("actionplan", ActionPlan.Type),
	}
}

// Mixin of the Risk
func (Risk) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		mixin.SoftDeleteMixin{},
		emixin.IDMixin{},
		emixin.TagMixin{},
	}
}

// Annotations of the Risk
func (Risk) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
	}
}