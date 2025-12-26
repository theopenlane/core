package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/common/models"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/entx/history"
)

// OrgModule holds module subscription info for an organization
// each enabled module is stored as a separate row
type OrgModule struct {
	SchemaFuncs

	ent.Schema
}

const SchemaOrgModule = "org_module"

// Name returns the name of the OrgModule schema
func (OrgModule) Name() string {
	return SchemaOrgModule
}

// GetType returns the type of the OrgModule schema
func (OrgModule) GetType() any {
	return OrgModule.Type
}

// PluralName returns the plural name of the OrgModule schema
func (OrgModule) PluralName() string {
	return pluralize.NewClient().Plural(SchemaOrgModule)
}

// Fields of the OrgModule
func (OrgModule) Fields() []ent.Field {
	return []ent.Field{
		field.String("module").
			GoType(models.OrgModule("")).
			NotEmpty().
			Comment("module key this record represents"),
		field.JSON("price", models.Price{}).
			Optional(),
		field.String("stripe_price_id").
			Optional(),
		field.String("status").
			Comment("the status of the module, e.g. active, inactive, trialing").
			Optional(),
		field.String("visibility").
			Optional(),
		field.Bool("active").
			Default(true),
		field.String("module_lookup_key").
			Comment("the lookup key for the module, used for Stripe integration").
			Optional(),
		field.String("subscription_id").
			Comment("the ID of the subscription this module is linked to").
			Optional(),
		field.String("price_id").
			Comment("the ID of the price associated with this module").
			Optional(),
	}
}

// Edges of the OrgModule
func (o OrgModule) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: o,
			t:          OrgSubscription.Type,
			edgeSchema: OrgSubscription{},
			field:      "subscription_id",
			ref:        "modules",
			annotations: []schema.Annotation{
				accessmap.EdgeNoAuthCheck(),
			},
		}),
		defaultEdgeToWithPagination(o, OrgProduct{}),
		defaultEdgeToWithPagination(o, OrgPrice{}),
	}
}

// Mixin returns the mixins for the OrgModule schema
func (o OrgModule) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeAnnotations: true,
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(o),
		},
	}.getMixins(o)
}

// Annotations returns the annotations for the OrgModule schema
func (OrgModule) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.Skip(entgql.SkipAll),
		entx.SchemaGenSkip(true),
		entx.QueryGenSkip(true),
		history.Annotations{
			Exclude: true,
		},
	}
}

// Hooks of the OrgModule
func (OrgModule) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookOrgModule(),
		hooks.HookOrgModuleUpdate(),
	}
}
