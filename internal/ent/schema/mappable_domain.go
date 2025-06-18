package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/validator"
	"github.com/theopenlane/entx"
)

// MappableDomain holds the schema definition for the MappableDomain entity
type MappableDomain struct {
	SchemaFuncs

	ent.Schema
}

const (
	// SchemaMappableDomain is the name of the MappableDomain schema.
	SchemaMappableDomain = "mappable_domain"
)

// Name returns the name of the schema
func (MappableDomain) Name() string {
	return SchemaMappableDomain
}

// GetType returns the type of the schema
func (MappableDomain) GetType() any {
	return MappableDomain.Type
}

// PluralName returns the plural name of the schema
func (MappableDomain) PluralName() string {
	return pluralize.NewClient().Plural(SchemaMappableDomain)
}

// Fields of the MappableDomain
func (MappableDomain) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("Name of the mappable domain").
			Validate(validator.ValidateURL()).
			MaxLen(maxDomainNameLen).
			NotEmpty().
			Immutable().
			Annotations(
				entgql.OrderField("name"),
			),
		field.String("zone_id").
			Comment("DNS Zone ID of the mappable domain.").
			NotEmpty().
			Immutable(),
	}
}

// Mixin of the MappableDomain
func (e MappableDomain) Mixin() []ent.Mixin {
	return mixinConfig{}.getMixins()
}

// Edges of the MappableDomain
func (e MappableDomain) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeToWithPagination(e, CustomDomain{}),
	}
}

// Indexes of the MappableDomain
func (MappableDomain) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name").
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Policy of the MappableDomain
func (MappableDomain) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(),
		),
		policy.WithMutationRules(
			rule.AllowMutationIfSystemAdmin(),
			privacy.AlwaysDenyRule(),
		),
	)
}

// Hooks of the MappableDomain
func (MappableDomain) Hooks() []ent.Hook {
	return []ent.Hook{}
}

// Annotations of the MappableDomain
func (MappableDomain) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entx.Features("trust-center"),
	}
}
