package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	emixin "github.com/theopenlane/entx/mixin"

	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/pkg/models"
)

// OrgSubscription holds the schema definition for the OrgSubscription entity
type OrgSubscription struct {
	ent.Schema
}

// Fields of the OrgSubscription
func (OrgSubscription) Fields() []ent.Field {
	return []ent.Field{
		field.String("stripe_subscription_id").
			Comment("the stripe subscription id").
			Optional(),
		field.String("product_tier").
			Comment("the common name of the product tier the subscription is associated with, e.g. starter tier").
			Optional(),
		field.JSON("product_price", models.Price{}).
			Comment("the price of the product tier").
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
			Nillable().
			Optional(),
		field.JSON("features", []string{}).
			Comment("the features associated with the subscription").
			Optional(),
		field.JSON("feature_lookup_keys", []string{}).
			Comment("the feature lookup keys associated with the subscription").
			Optional(),
	}
}

// Mixin of the OrgSubscription
func (OrgSubscription) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		emixin.IDMixin{},
		emixin.TagMixin{},
		mixin.SoftDeleteMixin{},
		NewOrgOwnMixinWithRef("org_subscriptions"),
	}
}

// Edges of the OrgSubscription
func (OrgSubscription) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the OrgSubscription
func (OrgSubscription) Indexes() []ent.Index {
	return []ent.Index{}
}

// Annotations of the OrgSubscription
func (OrgSubscription) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entgql.RelayConnection(),
		// since we only have queries, we can just use the interceptors for queries and can skip the fga generated checks
		// entfga.MembershipChecks("organization"),
	}
}

// Hooks of the OrgSubscription
func (OrgSubscription) Hooks() []ent.Hook {
	return []ent.Hook{}
}

// Interceptors of the OrgSubscription
func (OrgSubscription) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorSubscriptionURL(),
	}
}
