package schema

import (
	"context"

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
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/mixin"
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
		field.String("feature_id").Immutable().NotEmpty(),
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
		entfga.Annotations{
			ObjectType:      "organization",
			IncludeHooks:    false,
			NillableIDField: true,
			OrgOwnedField:   true,
			IDField:         "OwnerID",
		},
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
	return privacy.Policy{
		Mutation: privacy.MutationPolicy{
			privacy.EntitlementPlanFeatureMutationRuleFunc(func(ctx context.Context, m *generated.EntitlementPlanFeatureMutation) error {
				return m.CheckAccessForEdit(ctx)
			}),

			privacy.AlwaysDenyRule(),
		},
		Query: privacy.QueryPolicy{
			privacy.EntitlementPlanFeatureQueryRuleFunc(func(ctx context.Context, q *generated.EntitlementPlanFeatureQuery) error {
				return q.CheckAccess(ctx)
			}),
			privacy.AlwaysDenyRule(),
		},
	}
}
