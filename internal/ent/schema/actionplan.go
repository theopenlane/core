package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	emixin "github.com/theopenlane/entx/mixin"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/pkg/enums"
)

// ActionPlan defines the actionplan schema.
type ActionPlan struct {
	ent.Schema
}

// Fields returns actionplan fields.
func (ActionPlan) Fields() []ent.Field {
	return []ent.Field{
		field.Time("due_date").
			Optional().
			Comment("due date of the action plan"),
		field.Enum("priority").
			GoType(enums.Priority("")).
			Optional().
			Comment("priority of the action plan"),
		field.String("source").
			Optional().
			Comment("source of the action plan"),
	}
}

// Edges of the ActionPlan
func (ActionPlan) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("risk", Risk.Type).
			Ref("action_plans"),
		edge.From("control", Control.Type).
			Ref("action_plans"),
		edge.From("user", User.Type).
			Ref("action_plans"),
		edge.From("program", Program.Type).
			Ref("action_plans"),
	}
}

// Mixin of the ActionPlan
func (ActionPlan) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		mixin.SoftDeleteMixin{},
		emixin.IDMixin{},
		emixin.TagMixin{},

		DocumentMixin{DocumentType: "action plan"},
		mixin.RevisionMixin{},
	}
}

// Annotations of the ActionPlan
func (ActionPlan) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
	}
}

// Policy of the ActionPlan
func (ActionPlan) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysDenyRule(), // TODO(sfunk): - add query rules
		),
		policy.WithMutationRules(
			privacy.AlwaysDenyRule(), // TODO(sfunk): - add query rules
		),
	)
}
