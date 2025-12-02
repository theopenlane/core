package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/entx/history"
	"github.com/theopenlane/shared/models"
)

// OrgPrice represents a price attached to a subscription product
type OrgPrice struct {
	SchemaFuncs

	ent.Schema
}

const SchemaOrgPrice = "org_price"

// Name returns the name of the OrgPrice schema
func (OrgPrice) Name() string {
	return SchemaOrgPrice
}

// GetType returns the type of the OrgPrice schema
func (OrgPrice) GetType() any {
	return OrgPrice.Type
}

// PluralName returns the plural name of the OrgPrice schema
func (OrgPrice) PluralName() string {
	return pluralize.NewClient().Plural(SchemaOrgPrice)
}

// Fields of the OrgPrice
func (OrgPrice) Fields() []ent.Field {
	return []ent.Field{
		field.JSON("price", models.Price{}).
			Comment("the price details for this subscription product").
			Optional(),
		field.String("stripe_price_id").
			Comment("the Stripe price ID for this subscription product").
			Optional(),
		field.String("status").
			Comment("the status of the subscription product").
			Optional(),
		field.Bool("active").
			Default(true),
		field.String("product_id").
			Comment("the ID of the product this price is associated with").
			Optional(),
		field.String("subscription_id").
			Optional(),
	}
}

// Edges of the OrgPrice
func (o OrgPrice) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeFromWithPagination(o, OrgProduct{}),
		defaultEdgeFromWithPagination(o, OrgModule{}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: o,
			t:          OrgSubscription.Type,
			edgeSchema: OrgSubscription{},
			field:      "subscription_id",
			ref:        "prices",
			annotations: []schema.Annotation{
				accessmap.EdgeNoAuthCheck(),
			},
		}),
	}
}

// Mixin returns the mixins for the OrgPrice schema
func (o OrgPrice) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeAnnotations: true,
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(o),
		},
	}.getMixins(o)
}

// Annotations returns the annotations for the OrgPrice schema
func (OrgPrice) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.Skip(entgql.SkipAll),
		entx.SchemaGenSkip(true),
		entx.QueryGenSkip(true),
		history.Annotations{
			Exclude: true,
		},
	}
}

func (OrgPrice) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogBaseModule,
	}
}
