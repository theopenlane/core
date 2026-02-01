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

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/mixin"
)

// IntegrationOperationDefinition holds the schema definition for integration operations.
type IntegrationOperationDefinition struct {
	SchemaFuncs

	ent.Schema
}

// SchemaIntegrationOperationDefinition is the name of the IntegrationOperationDefinition schema.
const SchemaIntegrationOperationDefinition = "integration_operation_definition"

// Name returns the name of the IntegrationOperationDefinition schema.
func (IntegrationOperationDefinition) Name() string {
	return SchemaIntegrationOperationDefinition
}

// GetType returns the type of the IntegrationOperationDefinition schema.
func (IntegrationOperationDefinition) GetType() any {
	return IntegrationOperationDefinition.Type
}

// PluralName returns the plural name of the IntegrationOperationDefinition schema.
func (IntegrationOperationDefinition) PluralName() string {
	return pluralize.NewClient().Plural(SchemaIntegrationOperationDefinition)
}

// Fields of the IntegrationOperationDefinition.
func (IntegrationOperationDefinition) Fields() []ent.Field {
	return []ent.Field{
		field.String("integration_definition_id").
			Comment("integration definition this operation belongs to").
			NotEmpty(),
		field.String("name").
			Comment("stable operation identifier").
			NotEmpty().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("NAME"),
			),
		field.String("operation_kind").
			Comment("operation category (health_check, collect_findings, etc)").
			Optional().
			Annotations(
				entgql.OrderField("OPERATION_KIND"),
			),
		field.String("description").
			Comment("description of the operation").
			Optional(),
		field.JSON("config_schema", map[string]any{}).
			Comment("jsonschema for operation configuration").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.JSON("output_schema", map[string]any{}).
			Comment("jsonschema for operation output").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.JSON("default_config", map[string]any{}).
			Comment("default configuration payload").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.Strings("required_scopes").
			Comment("scopes required to execute the operation").
			Optional(),
		field.JSON("metadata", map[string]any{}).
			Comment("additional operation metadata").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.Bool("active").
			Comment("whether the operation is active").
			Default(true).
			Annotations(
				entgql.OrderField("ACTIVE"),
			),
	}
}

// Indexes of the IntegrationOperationDefinition.
func (IntegrationOperationDefinition) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("integration_definition_id", "name").
			Unique().
			Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Edges of the IntegrationOperationDefinition.
func (o IntegrationOperationDefinition) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: o,
			edgeSchema: IntegrationDefinition{},
			field:      "integration_definition_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(IntegrationDefinition{}.Name()),
			},
		}),
		defaultEdgeToWithPagination(o, IntegrationRun{}),
	}
}

// Mixin of the IntegrationOperationDefinition.
func (o IntegrationOperationDefinition) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags: true,
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[IntegrationOperationDefinition](o,
				withParents(IntegrationDefinition{}),
				withOrganizationOwner(true),
			),
			mixin.NewSystemOwnedMixin(),
		},
	}.getMixins(o)
}

// Modules of the IntegrationOperationDefinition.
func (IntegrationOperationDefinition) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogBaseModule,
	}
}

// Annotations of the IntegrationOperationDefinition.
func (IntegrationOperationDefinition) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.Skip(
			entgql.SkipMutationCreateInput,
			entgql.SkipMutationUpdateInput,
		),
	}
}

// Policy of the IntegrationOperationDefinition.
//func (IntegrationOperationDefinition) Policy() ent.Policy {
//	return policy.NewPolicy(
//		policy.WithQueryRules(
//			policy.CheckOrgReadAccess(),
//		),
//		policy.WithMutationRules(
//			rule.AllowMutationIfSystemAdmin(),
//			policy.CheckOrgWriteAccess(),
//		),
//	)
//}
//
