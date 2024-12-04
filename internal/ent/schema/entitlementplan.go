package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	emixin "github.com/theopenlane/entx/mixin"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
)

// EntitlementPlan holds the schema definition for the EntitlementPlan entity
type EntitlementPlan struct {
	ent.Schema
}

// Fields of the EntitlementPlan
func (EntitlementPlan) Fields() []ent.Field {
	return []ent.Field{
		field.String("display_name").
			Comment("the displayed 'friendly' name of the plan").
			Optional(),
		field.String("name").
			Comment("the unique name of the plan").
			Immutable().
			NotEmpty(),
		field.String("description").
			Comment("a description of the plan").
			Optional(),
		field.String("version").
			Comment("the version of the plan").
			NotEmpty().
			Immutable(),
		field.JSON("metadata", map[string]interface{}{}).
			Comment("metadata for the plan").
			Optional(),
		field.String("stripe_product_id").
			Comment("the product ID in Stripe").
			Optional(),
		field.String("stripe_price_id").
			Comment("the price ID in Stripe associated with the product").
			Optional(),
	}
}

// Mixin of the EntitlementPlan
func (EntitlementPlan) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		emixin.IDMixin{},
		mixin.SoftDeleteMixin{},
		emixin.TagMixin{},
		NewOrgOwnMixinWithRef("entitlementplans"),
	}
}

// Indexes of the EntitlementPlan
func (EntitlementPlan) Indexes() []ent.Index {
	return []ent.Index{
		// names should be unique, but ignore deleted names
		index.Fields("name", "version", "owner_id").
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Edges of the EntitlementPlan
func (EntitlementPlan) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("entitlements", Entitlement.Type),
		edge.From("base_features", Feature.Type).
			Ref("plans").
			Through("features", EntitlementPlanFeature.Type),
		edge.To("events", Event.Type),
	}
}

// Annotations of the EntitlementPlan
func (EntitlementPlan) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entgql.RelayConnection(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
		entfga.OrganizationInheritedChecks(),
	}
}

// Hooks of the EntitlementPlan
func (EntitlementPlan) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookEntitlementPlan(),
	}
}

// Policy of the EntitlementPlan
func (EntitlementPlan) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			entfga.CheckReadAccess[*generated.EntitlementPlanQuery](),
		),
		policy.WithMutationRules(
			entfga.CheckEditAccess[*generated.EntitlementPlanMutation](),
		),
	)
}
