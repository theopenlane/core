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

// OrgProduct represents a module product linked to a subscription.
type OrgProduct struct {
	SchemaFuncs

	ent.Schema
}

const SchemaOrgProduct = "org_product"

// Name returns the name of the OrgProduct schema
func (OrgProduct) Name() string {
	return SchemaOrgProduct
}

// GetType returns the type of the OrgProduct schema
func (OrgProduct) GetType() any {
	return OrgProduct.Type
}

// PluralName returns the plural name of the OrgProduct schema
func (OrgProduct) PluralName() string {
	return pluralize.NewClient().Plural(SchemaOrgProduct)
}

// Fields returns the fields for the OrgProduct schema
func (OrgProduct) Fields() []ent.Field {
	return []ent.Field{
		field.String("module").
			Comment("module key this product represents"),
		field.String("stripe_product_id").
			Comment("the Stripe product ID for this subscription product").
			Optional(),
		field.String("status").
			Comment("the status of the product, e.g. active, inactive, trialing").
			Optional(),
		field.Bool("active").
			Comment("indicates if the product is active or not").
			Default(true),
		field.String("subscription_id").
			Comment("the ID of the subscription this product is linked to").
			Optional(),
		field.String("price_id").
			Optional().
			Comment("the internal ID of the price this product is associated with"),
	}
}

// Edges returns the edges for the OrgProduct schema
func (o OrgProduct) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: o,
			edgeSchema: OrgSubscription{},
			field:      "subscription_id",
			ref:        "products",
			annotations: []schema.Annotation{
				accessmap.EdgeNoAuthCheck(),
			},
		}),
		defaultEdgeToWithPagination(o, OrgPrice{}),
		defaultEdgeToWithPagination(o, OrgModule{}),
	}
}

// Mixin returns the mixins for the OrgProduct schema
func (o OrgProduct) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeAnnotations: true,
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(o),
		},
	}.getMixins(o)
}

// Annotations returns the annotations for the OrgProduct schema
func (OrgProduct) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.Skip(entgql.SkipAll),
		entx.SchemaGenSkip(true),
		entx.QueryGenSkip(true),
		history.Annotations{
			Exclude: true,
		},
	}
}

func (OrgProduct) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogBaseModule,
	}
}
