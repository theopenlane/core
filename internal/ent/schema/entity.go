package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/validator"
)

// Entity holds the schema definition for the Entity entity
type Entity struct {
	SchemaFuncs

	ent.Schema
}

// SchemaEntity is the name of the Entity schema.
const SchemaEntity = "entity"

// Name returns the name of the Entity schema.
func (Entity) Name() string {
	return SchemaEntity
}

// GetType returns the type of the Entity schema.
func (Entity) GetType() any {
	return Entity.Type
}

// PluralName returns the plural name of the Entity schema.
func (Entity) PluralName() string {
	return pluralize.NewClient().Plural(SchemaEntity)
}

// Fields of the Entity
func (Entity) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name of the entity").
			Optional().
			SchemaType(map[string]string{
				dialect.Postgres: "citext",
			}).
			MinLen(minNameLength).
			Validate(validator.SpecialCharValidator).
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("name"),
			),
		field.String("display_name").
			Comment("The entity's displayed 'friendly' name").
			MaxLen(nameMaxLen).
			Optional().
			NotEmpty().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("display_name"),
			),
		field.String("description").
			Comment("An optional description of the entity").
			Optional().
			Annotations(
				entx.FieldSearchable(),
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.Strings("domains").
			Comment("domains associated with the entity").
			Validate(validator.ValidateDomains()).
			Optional(),
		field.String("entity_type_id").
			Comment("The type of the entity").
			Optional(),
		field.String("status").
			Comment("status of the entity").
			Default("active").
			Annotations(
				entgql.OrderField("status"),
			).
			Optional(),
	}
}

// Mixin of the Entity
func (e Entity) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(e),
			// TODO: this was added but there is no corresponding
			// fga type for entity so its not actually used
			// until that is added and the policy on the schema is updated
			newGroupPermissionsMixin(),
			mixin.NewSystemOwnedMixin(),
		},
	}.getMixins(e)
}

// Edges of the Entity
func (e Entity) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeToWithPagination(e, Contact{}),
		defaultEdgeToWithPagination(e, DocumentData{}),
		defaultEdgeToWithPagination(e, Note{}),
		defaultEdgeToWithPagination(e, File{}),
		defaultEdgeToWithPagination(e, Asset{}),
		defaultEdgeToWithPagination(e, Scan{}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: e,
			edgeSchema: EntityType{},
			field:      "entity_type_id",
		}),
	}
}

// Indexes of the Entity
func (Entity) Indexes() []ent.Index {
	return []ent.Index{
		// names should be unique, but ignore deleted names
		index.Fields("name", ownerFieldName).
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Hooks of the Entity
func (Entity) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookEntityCreate(),
	}
}

// Policy of the Entity
func (e Entity) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(),
		policy.WithMutationRules(
			policy.CheckOrgWriteAccess(),
		),
	)
}

func (Entity) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogEntityManagementModule,
	}
}
