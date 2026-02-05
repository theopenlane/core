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
	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
)

// Integration maps configured integrations (github, slack, etc.) to organizations
type Integration struct {
	SchemaFuncs

	ent.Schema
}

// SchemaIntegration is the name of the Integration schema.
const SchemaIntegration = "integration"

// Name returns the name of the Integration schema.
func (Integration) Name() string {
	return SchemaIntegration
}

// GetType returns the type of the Integration schema.
func (Integration) GetType() any {
	return Integration.Type
}

// PluralName returns the plural name of the Integration schema.
func (Integration) PluralName() string {
	return pluralize.NewClient().Plural(SchemaIntegration)
}

// Fields of the Integration
func (Integration) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name of the integration").
			NotEmpty().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("name"),
			),
		field.String("description").
			Comment("a description of the integration").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.String("kind").
			Comment("the kind of integration, such as github, slack, s3 etc.").
			Optional().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("kind"),
			),
		field.String("integration_type").
			Comment("the type of integration, such as communicattion, storage, SCM, etc.").
			Optional().
			Annotations(
				entgql.OrderField("integration_type"),
			),
		field.JSON("provider_metadata", openapi.IntegrationProviderMetadata{}).
			Comment("cached provider metadata for UI and registry access").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipType),
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.JSON("config", openapi.IntegrationConfig{}).
			Comment("runtime configuration for operations, scheduling, and mappings").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipType),
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.JSON("provider_state", openapi.IntegrationProviderState{}).
			Comment("provider-specific integration state captured during auth/config").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipType),
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.JSON("metadata", map[string]any{}).
			Comment("additional metadata about the integration").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
	}
}

// Edges of the Integration
func (i Integration) Edges() []ent.Edge {
	return []ent.Edge{
		edgeToWithPagination(&edgeDefinition{
			fromSchema: i,
			edgeSchema: Hush{},
			comment:    "the secrets associated with the integration",
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: i,
			edgeSchema: File{},
			comment:    "files associated with the integration",
		}),
		defaultEdgeToWithPagination(i, Event{}),
		defaultEdgeToWithPagination(i, Finding{}),
		defaultEdgeToWithPagination(i, Vulnerability{}),
		defaultEdgeToWithPagination(i, Review{}),
		defaultEdgeToWithPagination(i, Remediation{}),
		defaultEdgeToWithPagination(i, Task{}),
		defaultEdgeToWithPagination(i, ActionPlan{}),
		defaultEdgeToWithPagination(i, DirectoryAccount{}),
		defaultEdgeToWithPagination(i, DirectoryGroup{}),
		defaultEdgeToWithPagination(i, DirectoryMembership{}),
		defaultEdgeToWithPagination(i, DirectorySyncRun{}),
		defaultEdgeToWithPagination(i, NotificationTemplate{}),
		defaultEdgeToWithPagination(i, EmailTemplate{}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: i,
			edgeSchema: IntegrationWebhook{},
			annotations: []schema.Annotation{
				entgql.Skip(entgql.SkipAll),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: i,
			edgeSchema: IntegrationRun{},
			annotations: []schema.Annotation{
				entgql.Skip(entgql.SkipAll),
			},
		}),
		defaultEdgeFromWithPagination(i, Entity{}),
	}
}

// Indexes of the Integration
func (Integration) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields(ownerFieldName, "kind").
			Unique().
			Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Mixin of the Integration
func (i Integration) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(i),
			mixin.NewSystemOwnedMixin(mixin.SkipTupleCreation()),
			newCustomEnumMixin(i, withEnumFieldName("environment"), withGlobalEnum()),
			newCustomEnumMixin(i, withEnumFieldName("scope"), withGlobalEnum()),
		},
	}.getMixins(i)
}

// Policy of the Integration
func (i Integration) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckOrgWriteAccess(),
		),
	)
}

func (Integration) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogBaseModule,
	}
}

// Annotations of the Integration
func (Integration) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.Skip(
			// integrations are created by an oauth flow, not by the user directly
			entgql.SkipMutationCreateInput,
			entgql.SkipMutationUpdateInput,
		),
	}
}
