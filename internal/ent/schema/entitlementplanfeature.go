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
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
)

// EntitlementPlanFeature holds the schema definition for the EntitlementPlanFeature entity
type EntitlementPlanFeature struct {
	ent.Schema
}

// Fields of the EntitlementPlanFeature
func (EntitlementPlanFeature) Fields() []ent.Field {
	return []ent.Field{
		field.JSON("metadata", map[string]interface{}{}).
			Comment("metadata for the entitlement plan feature such as usage limits").
			Optional(),
		field.String("plan_id").Immutable().NotEmpty(),
		field.String("stripe_product_id").
			Comment("the product ID in Stripe").
			Optional(),
		field.String("feature_id").Immutable().NotEmpty(),
		field.String("stripe_feature_id").
			Comment("the feature ID in Stripe").
			Optional(),
	}
}

// Mixin of the EntitlementPlanFeature
func (EntitlementPlanFeature) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		emixin.IDMixin{},
		mixin.SoftDeleteMixin{},
		emixin.TagMixin{},
		NewOrgOwnMixinWithRef("entitlementplanfeatures"),
	}
}

// Edges of the EntitlementPlanFeature
func (EntitlementPlanFeature) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("plan", EntitlementPlan.Type).
			Field("plan_id").
			Required().
			Unique().
			Immutable(),
		edge.To("feature", Feature.Type).
			Field("feature_id").
			Required().
			Unique().
			Immutable(),
		edge.To("events", Event.Type),
	}
}

// Indexes of the EntitlementPlanFeature
func (EntitlementPlanFeature) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("feature_id", "plan_id").
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Annotations of the EntitlementPlanFeature
func (EntitlementPlanFeature) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entgql.RelayConnection(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
		entfga.OrganizationInheritedChecks(),
	}
}

// Hooks of the EntitlementPlanFeature
func (EntitlementPlanFeature) Hooks() []ent.Hook {
	return []ent.Hook{}
}

// Interceptors of the EntitlementPlanFeature
func (EntitlementPlanFeature) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}

// Policy of the EntitlementPlan
func (EntitlementPlanFeature) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			policy.CheckReadAccess[*generated.EntitlementPlanFeatureQuery](),
		),
		policy.WithMutationRules(
			policy.CheckEditAccess[*generated.EntitlementPlanFeatureMutation](),
		),
	)
}
