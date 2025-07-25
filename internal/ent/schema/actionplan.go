package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/pkg/enums"
)

// ActionPlan defines the actionplan schema.
type ActionPlan struct {
	SchemaFuncs

	ent.Schema
}

// SchemaActionPlan is the name of the actionplan schema.
const SchemaActionPlan = "action_plan"

// Name returns the name of the actionplan schema.
func (ActionPlan) Name() string {
	return SchemaActionPlan
}

// GetType returns the type of the actionplan schema.
func (ActionPlan) GetType() any {
	return ActionPlan.Type
}

// PluralName returns the plural name of the actionplan schema.
func (ActionPlan) PluralName() string {
	return pluralize.NewClient().Plural(SchemaActionPlan)
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
func (a ActionPlan) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeFromWithPagination(a, Risk{}),
		defaultEdgeFromWithPagination(a, Control{}),
		defaultEdgeFromWithPagination(a, User{}),
		defaultEdgeFromWithPagination(a, Program{}),
	}
}

// Mixin of the ActionPlan
func (a ActionPlan) Mixin() []ent.Mixin {
	return mixinConfig{
		includeRevision: true,
		additionalMixins: []ent.Mixin{
			NewDocumentMixin(a),
			newOrgOwnedMixin(a),
		}}.getMixins(a)
}

// Annotations of the ActionPlan
func (ActionPlan) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entx.Features("compliance", "continuous-compliance-automation"),
		entfga.SelfAccessChecks(),
	}
}

// Policy of the ActionPlan
func (ActionPlan) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.ActionPlanMutation](),
		),
	)
}
