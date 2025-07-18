package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/entx"
)

// OrgSubscription holds the schema definition for the OrgSubscription entity
type OrgSubscription struct {
	SchemaFuncs

	ent.Schema
}

// SchemaOrgSubscription is the name of the OrgSubscription schema.
const SchemaOrgSubscription = "org_subscription"

// Name returns the name of the OrgSubscription schema.
func (OrgSubscription) Name() string {
	return SchemaOrgSubscription
}

// GetType returns the type of the OrgSubscription schema.
func (OrgSubscription) GetType() any {
	return OrgSubscription.Type
}

// PluralName returns the plural name of the OrgSubscription schema.
func (OrgSubscription) PluralName() string {
	return pluralize.NewClient().Plural(SchemaOrgSubscription)
}

// Fields of the OrgSubscription
func (OrgSubscription) Fields() []ent.Field {
	return []ent.Field{
		field.String("stripe_subscription_id").
			Comment("the stripe subscription id").
			Optional(),
		field.String("product_tier").
			Comment("the common name of the product tier the subscription is associated with, e.g. starter tier").
			Annotations(
				entgql.OrderField("product_tier"),
			).
			Optional(),
		field.JSON("product_price", models.Price{}).
			Comment("the price of the product tier").
			Optional(),
		field.String("stripe_product_tier_id").
			Comment("the product id that represents the tier in stripe").
			Optional(),
		field.String("stripe_subscription_status").
			Comment("the status of the subscription in stripe -- see https://docs.stripe.com/api/subscriptions/object#subscription_object-status").
			Annotations(
				entgql.OrderField("stripe_subscription_status"),
			).
			Optional(),
		field.Bool("active").
			Comment("indicates if the subscription is active").
			Annotations(
				entgql.OrderField("active"),
			).
			Default(true),
		field.String("stripe_customer_id").
			Comment("the customer ID the subscription is associated to").
			Unique().
			Optional(),
		field.Time("expires_at").
			Comment("the time the subscription is set to expire; only populated if subscription is cancelled").
			Annotations(
				entgql.OrderField("expires_at"),
			).
			Nillable().
			Optional(),
		field.Time("trial_expires_at").
			Comment("the time the trial is set to expire").
			Annotations(
				entgql.OrderField("trial_expires_at"),
			).
			Nillable().
			Optional(),
		field.String("days_until_due").
			Comment("number of days until there is a due payment").
			Annotations(
				entgql.OrderField("days_until_due"),
			).
			Nillable().
			Optional(),
		field.Bool("payment_method_added").
			Comment("whether or not a payment method has been added to the account").
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
func (o OrgSubscription) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeAnnotations: true,
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(o,
				withSkipTokenTypesObjects(&token.SignUpToken{}),
				withHookFuncs(), // empty to skip the default hooks
			),
		},
	}.getMixins(o)
}

// Annotations of the OrgSubscription
func (OrgSubscription) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entx.Features("base"),
		entgql.QueryField(),
		entgql.RelayConnection(),
		// since we only have queries, we can just use the interceptors for queries and can skip the fga generated checks
	}
}

// Interceptors of the OrgSubscription
func (OrgSubscription) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorSubscriptionURL(),
	}
}

// Edges of the OrgSubscription
func (o OrgSubscription) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeToWithPagination(o, Event{}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: o,
			t:          OrgModule.Type,
			name:       "modules",
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: o,
			t:          OrgProduct.Type,
			name:       "products",
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: o,
			t:          OrgPrice.Type,
			name:       "prices",
		}),
	}
}
