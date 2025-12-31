package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/entx"
)

// Asset stores information about a discovered asset such as technology, domain, or device.
type Asset struct {
	SchemaFuncs
	ent.Schema
}

const SchemaAsset = "asset"

func (Asset) Name() string       { return SchemaAsset }
func (Asset) GetType() any       { return Asset.Type }
func (Asset) PluralName() string { return pluralize.NewClient().Plural(SchemaAsset) }

func (Asset) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("asset_type").
			Comment("the type of the asset, e.g. technology, domain, device, etc").
			GoType(enums.AssetType("")).
			Default(enums.AssetTypeTechnology.String()).
			Annotations(entgql.OrderField("ASSET_TYPE"), entx.FieldSearchable()),
		field.String("name").
			Comment("the name of the asset, e.g. matts computer, office router, IP address, etc").
			NotEmpty().
			Annotations(entgql.OrderField("name"), entx.FieldSearchable()),
		field.String("description").
			Optional(),
		field.String("identifier").
			Optional().
			Comment("unique identifier like domain, device id, etc"),
		field.String("website").
			Comment("the website of the asset, if applicable").
			Optional(),
		field.String("cpe").
			Comment("the CPE (Common Platform Enumeration) of the asset, if applicable").
			Optional().
			Annotations(entgql.Skip(entgql.SkipWhereInput)),
		field.Strings("categories").
			Comment("the categories of the asset, e.g. web server, database, etc").
			Optional(),
	}
}

func (a Asset) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.Asset](a,
				withOrganizationOwner(true),
			),
			newGroupPermissionsMixin(),
			mixin.NewSystemOwnedMixin(),
		},
	}.getMixins(a)
}

func (a Asset) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeFromWithPagination(a, Scan{}),
		defaultEdgeFromWithPagination(a, Entity{}),
		defaultEdgeFromWithPagination(a, Control{}),
	}
}

func (a Asset) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogEntityManagementModule,
	}
}

// Policy of the Asset
func (a Asset) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			rule.AllowMutationIfSystemAdmin(),
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.AssetMutation](),
		),
	)
}

// Annotations of the Asset
func (a Asset) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
	}
}

// Indexes of the Asset
func (Asset) Indexes() []ent.Index {
	return []ent.Index{
		// names should be unique, but ignore deleted names
		index.Fields("name", ownerFieldName).
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}
