package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/ent/mixin"
	"github.com/theopenlane/ent/privacy/policy"
	"github.com/theopenlane/ent/validator"
	"github.com/theopenlane/shared/models"
)

const (
	maxEntityNameLen = 64
)

// EntityType holds the schema definition for the EntityType entity
type EntityType struct {
	SchemaFuncs

	ent.Schema
}

// SchemaEntityType is the name of the EntityType schema.
const SchemaEntityType = "entity_type"

// Name returns the name of the EntityType schema.
func (EntityType) Name() string {
	return SchemaEntityType
}

// GetType returns the type of the EntityType schema.
func (EntityType) GetType() any {
	return EntityType.Type
}

// PluralName returns the plural name of the EntityType schema.
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
func (e EntityType) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(e),
			mixin.NewSystemOwnedMixin(),
		},
	}.getMixins(e)
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

// Policy of the EntityType
func (e EntityType) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(),
		policy.WithMutationRules(
			policy.CheckOrgWriteAccess(),
		),
	)
}

func (EntityType) Modules() []models.OrgModule {
	return []models.OrgModule{
		// entity types we seed for the user, so we can just include the base module
		models.CatalogBaseModule,
	}
}
