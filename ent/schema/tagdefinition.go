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
	"github.com/theopenlane/entx/history"

	"github.com/theopenlane/ent/hooks"
	"github.com/theopenlane/ent/mixin"
	"github.com/theopenlane/ent/privacy/policy"
	"github.com/theopenlane/ent/validator"
	"github.com/theopenlane/shared/models"
)

// TagDefinition holds the schema definition for the TagDefinition entity
type TagDefinition struct {
	SchemaFuncs

	ent.Schema
}

// SchemaTagDefinitionis the name of the schema in snake case
const SchemaTagDefinition = "tag_definition"

// Name is the name of the schema in snake case
func (TagDefinition) Name() string {
	return SchemaTagDefinition
}

// GetType returns the type of the schema
func (TagDefinition) GetType() any {
	return TagDefinition.Type
}

// PluralName returns the plural name of the schema
func (TagDefinition) PluralName() string {
	return pluralize.NewClient().Plural(SchemaTagDefinition)
}

// Fields of the TagDefinition
func (TagDefinition) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("The name of the tag definition").
			Immutable().
			Annotations(entx.FieldSearchable()).
			SchemaType(map[string]string{
				dialect.Postgres: "citext",
			}).
			NotEmpty(),
		field.Strings("aliases").
			Comment("common aliases or misspellings for the tag definition").
			Optional(),
		field.String("slug").
			Comment("The slug of the tag definition, derived from the name, unique per organization").
			NotEmpty().
			SchemaType(map[string]string{
				dialect.Postgres: "citext",
			}).
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			).
			Optional(),
		field.String("description").
			Comment("The description of the tag definition").
			Optional(),
		field.String("color").
			Comment("The color of the tag definition in hex format").
			Validate(validator.HexColorValidator).
			DefaultFunc(defaultRandomColor).
			Optional(),
	}
}

// Mixin of the TagDefinition
func (t TagDefinition) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags: true,
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(t),
			mixin.NewSystemOwnedMixin(),
		},
	}.getMixins(t)
}

// Edges of the TagDefinition
func (t TagDefinition) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the TagDefinition
func (TagDefinition) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("slug", ownerFieldName).
			Unique().
			Annotations(entsql.IndexWhere("deleted_at is NULL")),
		index.Fields("name", ownerFieldName).
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Annotations of the TagDefinition
func (TagDefinition) Annotations() []schema.Annotation {
	return []schema.Annotation{
		history.Annotations{
			Exclude: true,
		},
	}
}

// Hooks of the TagDefinition
func (TagDefinition) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookTagDefintion(),
		hooks.HookTagDefinitionDelete(),
	}
}

// Interceptors of the TagDefinition
func (TagDefinition) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}

// Modules this schema has access to
func (TagDefinition) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogBaseModule,
	}
}

// Policy of the TagDefinition
func (TagDefinition) Policy() ent.Policy {
	// add the new policy here, the default post-policy is to deny all
	// so you need to ensure there are rules in place to allow the actions you want
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckOrgWriteAccess(),
		),
	)
}
