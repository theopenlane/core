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

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/validator"
)

// IntegrationDefinition holds the schema definition for integration provider definitions.
type IntegrationDefinition struct {
	SchemaFuncs

	ent.Schema
}

// SchemaIntegrationDefinition is the name of the IntegrationDefinition schema.
const SchemaIntegrationDefinition = "integration_definition"

// Name returns the name of the IntegrationDefinition schema.
func (IntegrationDefinition) Name() string {
	return SchemaIntegrationDefinition
}

// GetType returns the type of the IntegrationDefinition schema.
func (IntegrationDefinition) GetType() any {
	return IntegrationDefinition.Type
}

// PluralName returns the plural name of the IntegrationDefinition schema.
func (IntegrationDefinition) PluralName() string {
	return pluralize.NewClient().Plural(SchemaIntegrationDefinition)
}

// Fields of the IntegrationDefinition.
func (IntegrationDefinition) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").
			Comment("stable provider identifier, e.g. github, slack").
			NotEmpty().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("KEY"),
			),
		field.String("display_name").
			Comment("user-facing provider name").
			Optional().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("DISPLAY_NAME"),
			),
		field.String("description").
			Comment("description of the provider").
			Optional(),
		field.String("category").
			Comment("provider category (code, collab, etc)").
			Optional().
			Annotations(
				entgql.OrderField("CATEGORY"),
			),
		field.String("auth_type").
			Comment("authentication type for the provider").
			Optional().
			Annotations(
				entgql.OrderField("AUTH_TYPE"),
			),
		field.String("docs_url").
			Comment("documentation URL for the provider").
			Optional().
			Validate(validator.ValidateURL()).
			Nillable(),
		field.String("logo_url").
			Comment("logo URL for the provider").
			Optional().
			Validate(validator.ValidateURL()).
			Nillable(),
		field.Bool("active").
			Comment("whether the provider is active").
			Default(true).
			Annotations(
				entgql.OrderField("ACTIVE"),
			),
		field.JSON("labels", map[string]string{}).
			Comment("provider labels for UI metadata").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.JSON("credential_schema", map[string]any{}).
			Comment("credential schema for the provider").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.JSON("metadata", map[string]any{}).
			Comment("additional provider metadata").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.JSON("provider_config", map[string]any{}).
			Comment("raw provider configuration payload").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
	}
}

// Indexes of the IntegrationDefinition.
func (IntegrationDefinition) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields(ownerFieldName, "key").
			Unique().
			Annotations(entsql.IndexWhere("deleted_at is NULL")),
		index.Fields("key").
			Unique().
			Annotations(entsql.IndexWhere("deleted_at is NULL and system_owned = true")),
	}
}

// Edges of the IntegrationDefinition.
func (i IntegrationDefinition) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeToWithPagination(i, Integration{}),
		defaultEdgeToWithPagination(i, IntegrationOperationDefinition{}),
	}
}

// Mixin of the IntegrationDefinition.
func (i IntegrationDefinition) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags: true,
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[IntegrationDefinition](i,
				withOrganizationOwner(true),
			),
			mixin.NewSystemOwnedMixin(),
		},
	}.getMixins(i)
}

// Modules of the IntegrationDefinition.
func (IntegrationDefinition) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogBaseModule,
	}
}

// Annotations of the IntegrationDefinition.
func (IntegrationDefinition) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.Skip(
			entgql.SkipMutationCreateInput,
			entgql.SkipMutationUpdateInput,
		),
	}
}

// Policy of the IntegrationDefinition.
func (IntegrationDefinition) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			policy.CheckOrgReadAccess(),
		),
		policy.WithMutationRules(
			rule.AllowMutationIfSystemAdmin(),
			policy.CheckOrgWriteAccess(),
		),
	)
}
