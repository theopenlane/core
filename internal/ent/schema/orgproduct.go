package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/history"
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
			Optional(),
		field.String("status").
			Optional(),
		field.Bool("active").
			Default(true),
		field.Time("trial_expires_at").
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("trial_expires_at"),
			),
		field.Time("expires_at").
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("expires_at"),
			),
		field.String("subscription_id").
			Optional(),
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
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: o,
			t:          OrgPrice.Type,
			name:       "prices",
		}),
	}
}

// Mixin returns the mixins for the OrgProduct schema
func (o OrgProduct) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeAnnotations: true,
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(o),
		},
	}.getMixins()
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
