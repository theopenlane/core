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
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/validator"
)

// Entity holds the schema definition for the Entity entity
type Entity struct {
	CustomSchema

	ent.Schema
}

const SchemaEntity = "entity"

func (Entity) Name() string {
	return SchemaEntity
}

func (Entity) GetType() any {
	return Entity.Type
}

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
			MinLen(3).
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
func (Entity) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			NewOrgOwnMixinWithRef("entities"),
		},
	}.getMixins()
}

// Edges of the Entity
func (e Entity) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeToWithPagination(e, Contact{}),
		defaultEdgeToWithPagination(e, DocumentData{}),
		defaultEdgeToWithPagination(e, Note{}),
		defaultEdgeToWithPagination(e, File{}),
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

// Annotations of the Entity
func (Entity) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.OrganizationInheritedChecks(),
	}
}

// Hooks of the Entity
func (Entity) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookEntityCreate(),
	}
}

// Policy of the Entity
func (Entity) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(), //  interceptor should filter out the results
		),
		policy.WithMutationRules(
			entfga.CheckEditAccess[*generated.EntityMutation](),
		),
	)
}
