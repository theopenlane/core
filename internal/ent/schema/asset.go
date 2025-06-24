package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/pkg/enums"
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
			GoType(enums.AssetType("")).
			Default(enums.AssetTypeTechnology.String()).
			Annotations(entgql.OrderField("asset_type")),
		field.String("name").NotEmpty().Annotations(entgql.OrderField("name")),
		field.String("description").Optional(),
		field.String("identifier").Optional().
			Comment("unique identifier like domain, device id, etc"),
		field.String("website").Optional(),
		field.String("cpe").Optional().Annotations(entgql.Skip(entgql.SkipWhereInput)),
		field.Strings("categories").Optional(),
	}
}

func (a Asset) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(a),
		},
	}.getMixins()
}

func (a Asset) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeFromWithPagination(a, Scan{}),
		defaultEdgeFromWithPagination(a, Risk{}),
	}
}

func (Asset) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			rule.AllowQueryIfSystemAdmin(),
			// object owner check done via scan
			policy.CheckOrgReadAccess(),
		),
		policy.WithMutationRules(
			rule.AllowMutationIfSystemAdmin(),
			policy.CheckOrgWriteAccess(),
		),
	)
}
