package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/theopenlane/entx/history"
	emixin "github.com/theopenlane/entx/mixin"
	"github.com/theopenlane/iam/entfga"

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
			Optional().
			StructTag(`json:"features"`),
	}
}

// Mixin of the Subscription
func (Subscription) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		emixin.IDMixin{},
		emixin.TagMixin{},
		mixin.SoftDeleteMixin{},
		NewOrgOwnedMixin(ObjectOwnedMixin{
			Ref: "subscriptions",
		}),
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
		entfga.Annotations{
			ObjectType:   "organization",
			IncludeHooks: false,
			IDField:      "OwnerID",
		},
		history.Annotations{
			Exclude: false,
		},
	}
}
