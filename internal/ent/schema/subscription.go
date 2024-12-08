package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	emixin "github.com/theopenlane/entx/mixin"

	"github.com/theopenlane/core/internal/ent/mixin"
)

// Subscription holds the schema definition for the Subscription entity
type Subscription struct {
	ent.Schema
}

// Fields of the Subscription
func (Subscription) Fields() []ent.Field {
	return []ent.Field{
		field.String("stripe_subscription_id").
			Comment("the stripe subscription id").
			Optional(),
		field.String("product_tier").
			Comment("the common name of the product tier the subscription is associated with, e.g. starter tier").
			Optional(),
		field.String("stripe_product_tier_id").
			Comment("the product id that represents the tier in stripe").
			Optional(),
		field.String("stripe_subscription_status").
			Comment("the status of the subscription in stripe -- see https://docs.stripe.com/api/subscriptions/object#subscription_object-status").
			Optional(),
		field.Bool("active").
			Comment("indicates if the subscription is active").
			Default(true),
		field.String("stripe_customer_id").
			Comment("the customer ID the subscription is associated to").
			Unique().
			Optional(),
		field.Time("expires_at").
			Comment("the time the subscription is set to expire; only populated if subscription is cancelled").
			Optional(),
		field.JSON("features", []string{}).
			Comment("the features associated with the subscription").
			Optional(),
	}
}

// Mixin of the Subscription
func (Subscription) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		emixin.IDMixin{},
		emixin.TagMixin{},
		mixin.SoftDeleteMixin{},
		NewOrgOwnMixinWithRef("subscriptions"),
	}
}

// Edges of the Subscription
func (Subscription) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the Subscription
func (Subscription) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("features", "owner_id"),
	}
}

// Annotations of the Subscription
func (Subscription) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entgql.RelayConnection(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
	}
}

// Hooks of the Subscription
func (Subscription) Hooks() []ent.Hook {
	return []ent.Hook{}
}

// Interceptors of the Subscription
func (Subscription) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}

// Policy of the Subscription
//func (Subscription) Policy() ent.Policy {
//	return policy.NewPolicy(
//		policy.WithQueryRules(
//			entfga.CheckReadAccess[*generated.Subscriptio](),
//		),
//		policy.WithMutationRules(
//			entfga.CheckEditAccess[*generated.ContactMutation](),
//		),
//	)
//}
//
