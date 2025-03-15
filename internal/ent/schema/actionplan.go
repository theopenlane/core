package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	emixin "github.com/theopenlane/entx/mixin"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
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
			Annotations(
				entgql.OrderField("due_date"),
			).
			Comment("due date of the action plan"),
		field.Enum("priority").
			GoType(enums.Priority("")).
			Annotations(
				entgql.OrderField("PRIORITY"),
			).
			Optional().
			Comment("priority of the action plan"),
		field.String("source").
			Annotations(
				entgql.OrderField("source"),
			).
			Optional().
			Comment("source of the action plan"),
	}
}

// Edges of the ActionPlan
func (ActionPlan) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("risk", Risk.Type).
			Annotations(entgql.RelayConnection()).
			Ref("action_plans"),
		edge.From("control", Control.Type).
			Annotations(entgql.RelayConnection()).
			Ref("action_plans"),
		edge.From("user", User.Type).
			Annotations(entgql.RelayConnection()).
			Ref("action_plans"),
		edge.From("program", Program.Type).
			Annotations(entgql.RelayConnection()).
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

		DocumentMixin{DocumentType: "action_plan"},
		mixin.RevisionMixin{},
		// all action plans must be associated to an organization
		NewOrgOwnMixinWithRef("action_plans"),
	}
}

// Annotations of the ActionPlan
func (ActionPlan) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.MultiOrder(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
		entfga.SelfAccessChecks(),
	}
}

// Policy of the ActionPlan
func (ActionPlan) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(), //  interceptor should filter out the results
		),
		policy.WithMutationRules(
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.ActionPlanMutation](),
		),
	)
}
