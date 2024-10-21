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
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/mixin"
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
		entfga.Annotations{
			ObjectType:      "organization",
			IncludeHooks:    false,
			NillableIDField: true,
			OrgOwnedField:   true,
			IDField:         "OwnerID",
		},
	}
}

// Hooks of the EntitlementPlan
func (EntitlementPlan) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookEntitlementPlan(),
	}
}

// Interceptors of the EntitlementPlan
func (EntitlementPlan) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}

// Policy of the EntitlementPlan
func (EntitlementPlan) Policy() ent.Policy {
	return privacy.Policy{
		Mutation: privacy.MutationPolicy{
			privacy.EntitlementPlanMutationRuleFunc(func(ctx context.Context, m *generated.EntitlementPlanMutation) error {
				return m.CheckAccessForEdit(ctx)
			}),

			privacy.AlwaysDenyRule(),
		},
		Query: privacy.QueryPolicy{
			privacy.EntitlementPlanQueryRuleFunc(func(ctx context.Context, q *generated.EntitlementPlanQuery) error {
				return q.CheckAccess(ctx)
			}),
			privacy.AlwaysDenyRule(),
		},
	}
}
