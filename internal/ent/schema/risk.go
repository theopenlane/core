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
			Comment("description of the risk"),
		field.String("status").
			Comment("status of the risk - mitigated or not, inflight, etc."),
		field.String("type").
			Comment("type of the risk, e.g. strategic, operational, financial, external, etc."),
		field.Text("business costs").
			Comment("business costs associated with the risk"),
		field.Text("impact").
			Comment("impact of the risk"),
		field.Text("likelihood").
			Comment("likelihood of the risk"),
		field.Text("mitigation").
			Comment("mitigation of the risk"),
		field.Text("satisfies").
			Comment("which controls are satisfied by the risk"),
		field.Text("severity").
			Comment("severity of the risk, e.g. high medium low"),
	}
}

// Edges of the Risk
func (Risk) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("control", User.Type).
			Ref("controls"),
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
