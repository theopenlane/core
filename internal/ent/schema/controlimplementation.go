package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	emixin "github.com/theopenlane/entx/mixin"
)

// ControlImplementation holds the schema definition for the ControlImplementation entity
type ControlImplementation struct {
	ent.Schema
}

// Fields of the ControlImplementation
func (ControlImplementation) Fields() []ent.Field {
	return []ent.Field{
		field.String("status").
			Optional().
			Comment("status of the control implementation"),
		field.Time("implementation_date").
			Optional().
			Comment("date the control was implemented"),
		field.Bool("verified").
			Optional().
			Comment("set to true if the control implementation has been verified"),
		field.Time("verification_date").
			Optional().
			Comment("date the control implementation was verified"),
		field.Text("details").
			Optional().
			Comment("details of the control implementation"),
	}
}

// Mixin of the ControlImplementation
func (ControlImplementation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		emixin.IDMixin{},
		mixin.SoftDeleteMixin{},
		emixin.TagMixin{},
	}
}

// Edges of the ControlImplementation
func (ControlImplementation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("controls", Control.Type).
			Ref("control_implementations"),
	}
}

// Annotations of the ControlImplementation
func (ControlImplementation) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entgql.RelayConnection(),
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
	}
}

// Policy of the ControlImplementation
func (ControlImplementation) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysDenyRule(), // TODO(sfunk): - add query rules
		),
		policy.WithMutationRules(
			privacy.AlwaysDenyRule(), // TODO(sfunk): - add query rules
		),
	)
}
