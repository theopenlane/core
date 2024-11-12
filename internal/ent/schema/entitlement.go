package schema

import (
	"context"
	"time" // Add this import

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

// Entitlement holds the schema definition for the Entitlement entity.
type Entitlement struct {
	ent.Schema
}

// Fields of the Entitlement.
func (Entitlement) Fields() []ent.Field {
	return []ent.Field{
		field.String("plan_id").
			Comment("the plan to which the entitlement belongs").
			NotEmpty().
			Immutable(),
		field.String("organization_id").
			Comment("the organization to which the entitlement belongs").
			NotEmpty().
			Immutable(),
		field.String("external_customer_id").
			Comment("used to store references to external systems, e.g. Stripe").
			Optional(),
		field.String("external_subscription_id").
			Comment("used to store references to external systems, e.g. Stripe").
			Optional(),
		field.Bool("expires").
			Comment("whether or not the customers entitlement expires - expires_at will show the time").
			Annotations(
				entgql.Skip(
					// skip these fields in the mutation
					// it will automatically be set based on the value of expires_at
					entgql.SkipMutationCreateInput,
					entgql.SkipMutationUpdateInput,
				),
			).
			Default(true),
		field.Time("expires_at").
			Comment("the time at which a customer's entitlement will expire, e.g. they've cancelled but paid through the end of the month").
			Optional().
			Nillable(),
		field.Bool("cancelled").
			Comment("whether or not the customer has cancelled their entitlement - usually used in conjunction with expires and expires at").
			Default(false),
		field.Time("cancelled_date").
			Comment("the date at which the customer cancelled their entitlement").
			Optional(),
		field.Time("bill_starting").
			Comment("the date at which the customer's billing starts").
			Default(func() time.Time {
				return time.Now().AddDate(0, 0, 14) // Set the date 14 days from when the entitlement is created
			}),
		field.Bool("active").
			Comment("whether or not the entitlement is active").
			Default(true),
	}
}

// Indexes of the Entitlement
func (Entitlement) Indexes() []ent.Index {
	return []ent.Index{
		// organizations should only have one active entitlement
		index.Fields("organization_id", "owner_id").
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL and cancelled = false")),
	}
}

// Edges of the Entitlement
func (Entitlement) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("plan", EntitlementPlan.Type).
			Field("plan_id").
			Unique().
			Required().
			Immutable().
			Ref("entitlements"),
		// Organization that is assigned the entitlement
		edge.From("organization", Organization.Type).
			Ref("organization_entitlement").
			Field("organization_id").
			Required().
			Immutable().
			Unique(),
		edge.To("events", Event.Type),
	}
}

// Annotations of the Entitlement
func (Entitlement) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
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

// Mixin of the Entitlement
func (Entitlement) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		emixin.IDMixin{},
		emixin.TagMixin{},
		mixin.SoftDeleteMixin{},
		NewOrgOwnMixinWithRef("entitlements"),
	}
}

// Hooks of the Entitlement
func (Entitlement) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookEntitlement(),
	}
}

// Policy of the Entitlement
func (Entitlement) Policy() ent.Policy {
	return privacy.Policy{
		Mutation: privacy.MutationPolicy{
			privacy.EntitlementMutationRuleFunc(func(ctx context.Context, m *generated.EntitlementMutation) error {
				return m.CheckAccessForEdit(ctx)
			}),

			privacy.AlwaysDenyRule(),
		},
		Query: privacy.QueryPolicy{
			privacy.EntitlementQueryRuleFunc(func(ctx context.Context, q *generated.EntitlementQuery) error {
				return q.CheckAccess(ctx)
			}),
			privacy.AlwaysDenyRule(),
		},
	}
}
