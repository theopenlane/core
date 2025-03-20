package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/validator"
)

const (
	maxEntityNameLen = 64
)

// EntityType holds the schema definition for the EntityType entity
type EntityType struct {
	CustomSchema

	ent.Schema
}

const SchemaEntityType = "entity_type"

func (EntityType) Name() string {
	return SchemaEntityType
}

func (EntityType) GetType() any {
	return EntityType.Type
}

func (EntityType) PluralName() string {
	return pluralize.NewClient().Plural(SchemaEntityType)
}

// Fields of the EntityType
func (EntityType) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name of the entity").
			SchemaType(map[string]string{
				dialect.Postgres: "citext",
			}).
			Validate(validator.SpecialCharValidator).
			MaxLen(maxEntityNameLen).
			NotEmpty().
			Annotations(
				entgql.OrderField("name"),
			),
	}
}

// Mixin of the EntityType
func (EntityType) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			NewOrgOwnMixinWithRef("entity_types"),
		},
	}.getMixins()
}

// Edges of the EntityType
func (e EntityType) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeToWithPagination(e, Entity{}),
	}
}

// Indexes of the EntityType
func (EntityType) Indexes() []ent.Index {
	return []ent.Index{
		// names should be unique by owner, but ignore deleted names
		index.Fields("name", ownerFieldName).
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Annotations of the EntityType
func (EntityType) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.OrganizationInheritedChecks(),
	}
}

// Policy of the EntityType
func (EntityType) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(), //  interceptor should filter out the results
		),
		policy.WithMutationRules(
			entfga.CheckEditAccess[*generated.EntityTypeMutation](),
		),
	)
}
