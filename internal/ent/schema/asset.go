package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
)

// Asset stores information about a discovered asset such as technology, domain, or device.
type Asset struct {
	SchemaFuncs
	ent.Schema
}

// SchemaAsset is the name of the Asset schema
const SchemaAsset = "asset"

// Name returns the name of the Asset schema
func (Asset) Name() string {
	return SchemaAsset
}

// GetType returns the type of the Asset schema
func (Asset) GetType() any {
	return Asset.Type
}

// PluralName returns the plural name of the Asset schema
func (Asset) PluralName() string {
	return pluralize.NewClient().Plural(SchemaAsset)
}

// Fields of the Asset
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
		field.String("physical_location").
			Comment("physical location of the asset, if applicable").
			Optional().
			Annotations(
				entgql.OrderField("physical_location"),
			),
		field.String("region").
			Comment("the region where the asset operates or is hosted").
			Optional().
			Annotations(
				entgql.OrderField("region"),
			),
		field.Bool("contains_pii").
			Comment("whether the asset stores or processes PII").
			Default(false).
			Optional().
			Annotations(
				entgql.OrderField("contains_pii"),
			),
		field.Enum("source_type").
			Comment("the source of the asset record, e.g., manual, discovered, imported, api").
			GoType(enums.SourceType("")).
			Default(enums.SourceTypeManual.String()).
			Annotations(
				entgql.OrderField("SOURCE_TYPE"),
			),
		field.String("source_platform_id").
			Comment("the platform that sourced the asset record").
			Optional(),
		field.String("source_identifier").
			Comment("the identifier used by the source platform for the asset").
			Optional().
			Annotations(
				entgql.OrderField("source_identifier"),
			),
		field.String("cost_center").
			Comment("cost center associated with the asset").
			Optional().
			Annotations(
				entgql.OrderField("cost_center"),
			),
		field.Float("estimated_monthly_cost").
			Comment("estimated monthly cost for the asset").
			Optional().
			Annotations(
				entgql.OrderField("estimated_monthly_cost"),
			),
		field.Time("purchase_date").
			Comment("purchase date for the asset").
			GoType(models.DateTime{}).
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("purchase_date"),
			),
		field.String("cpe").
			Comment("the CPE (Common Platform Enumeration) of the asset, if applicable").
			Optional().
			Annotations(entgql.Skip(entgql.SkipWhereInput)),
		field.Strings("categories").
			Comment("the categories of the asset, e.g. web server, database, etc").
			Optional(),
	}
}

// Mixin of the Asset
func (a Asset) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.Asset](a,
				withParents(Organization{}, Platform{}),
				withOrganizationOwner(true),
			),
			newGroupPermissionsMixin(),
			newResponsibilityMixin(a, withInternalOwner()),
			newCustomEnumMixin(a, withEnumFieldName("subtype")),
			newCustomEnumMixin(a, withEnumFieldName("data_classification")),
			newCustomEnumMixin(a, withEnumFieldName("environment"), withGlobalEnum()),
			newCustomEnumMixin(a, withEnumFieldName("scope"), withGlobalEnum()),
			newCustomEnumMixin(a, withEnumFieldName("access_model"), withGlobalEnum()),
			newCustomEnumMixin(a, withEnumFieldName("encryption_status"), withGlobalEnum()),
			newCustomEnumMixin(a, withEnumFieldName("security_tier"), withGlobalEnum()),
			newCustomEnumMixin(a, withEnumFieldName("criticality"), withGlobalEnum()),
			mixin.NewSystemOwnedMixin(mixin.SkipTupleCreation()),
		},
	}.getMixins(a)
}

// Edges of the Asset
func (a Asset) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeFromWithPagination(a, Scan{}),
		defaultEdgeFromWithPagination(a, Entity{}),
		defaultEdgeFromWithPagination(a, Platform{}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: a,
			name:       "out_of_scope_platforms",
			t:          Platform.Type,
			ref:        "out_of_scope_assets",
		}),
		defaultEdgeFromWithPagination(a, IdentityHolder{}),
		defaultEdgeFromWithPagination(a, Control{}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: a,
			name:       "source_platform",
			t:          Platform.Type,
			field:      "source_platform_id",
			ref:        "source_assets",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Platform{}.Name()),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: a,
			t:          Asset.Type,
			name:       "connected_assets",
			comment:    "assets that this asset connects to",
		}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: a,
			t:          Asset.Type,
			name:       "connected_from",
			comment:    "assets that connect to this asset",
			ref:        "connected_assets",
		}),
	}
}

// Modules this schema has access to
func (a Asset) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogEntityManagementModule,
		models.CatalogComplianceModule,
	}
}

// Policy of the Asset
func (a Asset) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			rule.AllowMutationIfSystemAdmin(),
			policy.CheckCreateAccess(),
			policy.CanCreateObjectsUnderParents([]string{
				Platform{}.PluralName(),
			}),
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
