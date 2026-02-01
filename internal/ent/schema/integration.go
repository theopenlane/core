package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/common/models"
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
				entgql.OrderField("kind"),
			),
		field.String("integration_type").
			Comment("the type of integration, such as communicattion, storage, SCM, etc.").
			Optional().
			Annotations(
				entgql.OrderField("integration_type"),
			),
		field.String("integration_definition_id").
			Comment("integration definition backing this connection").
			Optional(),
		field.String("connection_name").
			Comment("optional connection label to allow multiple connections per provider").
			Optional().
			Annotations(
				entgql.OrderField("CONNECTION_NAME"),
			),
		field.Bool("is_primary").
			Comment("whether this is the primary connection for the provider").
			Default(false),
		field.String("installation_id").
			Comment("external installation identifier for webhook lookups").
			Optional(),
		field.String("app_id").
			Comment("external application identifier for webhook lookups").
			Optional(),
		field.String("tenant_id").
			Comment("external tenant identifier for webhook lookups").
			Optional(),
		field.String("account_id").
			Comment("external account identifier for webhook lookups").
			Optional(),
		field.String("project_id").
			Comment("external project identifier for webhook lookups").
			Optional(),
		field.String("region").
			Comment("external region identifier for webhook lookups").
			Optional(),
		field.String("status").
			Comment("connection status for the integration").
			Optional().
			Annotations(
				entgql.OrderField("status"),
			),
		field.Time("last_health_check_at").
			Comment("last time a health check was performed").
			GoType(models.DateTime{}).
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("last_health_check_at"),
			),
		field.String("last_health_check_status").
			Comment("status of the last health check").
			Optional().
			Annotations(
				entgql.OrderField("last_health_check_status"),
			),
		field.Text("last_health_check_error").
			Comment("error details from the last health check").
			Optional(),
		field.String("auth_type").
			Comment("authentication type for the integration (oauth2, apikey, workload_identity, etc.)").
			Optional().
			Annotations(
				entgql.OrderField("auth_type"),
			),
		field.Strings("scopes").
			Comment("scopes granted for the integration").
			Optional(),
		field.Time("last_run_at").
			Comment("last time an integration action ran").
			GoType(models.DateTime{}).
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("last_run_at"),
			),
		field.String("last_run_status").
			Comment("status of the last integration run").
			Optional().
			Annotations(
				entgql.OrderField("last_run_status"),
			),
		field.Text("last_run_summary").
			Comment("summary of the last integration run").
			Optional(),
		field.JSON("last_run_details", map[string]any{}).
			Comment("structured details for the last integration run").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.JSON("available_actions", []map[string]any{}).
			Comment("available actions/operations published by the provider").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.JSON("provider_config", map[string]any{}).
			Comment("provider configuration payload for this integration").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.JSON("credential_schema", map[string]any{}).
			Comment("credential schema for provider configuration").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.JSON("operation_schemas", map[string]any{}).
			Comment("operation config/output schemas for available actions").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.JSON("instructions", map[string]any{}).
			Comment("setup or usage instructions for this integration").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.JSON("metadata", map[string]any{}).
			Comment("additional metadata about the integration").
			Optional(),
	}
}

// Edges of the Integration
func (i Integration) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: i,
			edgeSchema: IntegrationDefinition{},
			field:      "integration_definition_id",
		}),
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
		defaultEdgeToWithPagination(i, IntegrationWebhook{}),
		defaultEdgeToWithPagination(i, IntegrationRun{}),
		defaultEdgeFromWithPagination(i, Entity{}),
	}
}

// Indexes of the Integration
func (Integration) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields(ownerFieldName, "kind").
			Unique().
			Annotations(entsql.IndexWhere("deleted_at is NULL and is_primary = true")),
		index.Fields(ownerFieldName, "kind", "connection_name").
			Unique().
			Annotations(entsql.IndexWhere("deleted_at is NULL and connection_name is not NULL")),
		index.Fields("kind", "installation_id").
			Annotations(entsql.IndexWhere("deleted_at is NULL and installation_id is not NULL")),
		index.Fields("kind", "app_id").
			Annotations(entsql.IndexWhere("deleted_at is NULL and app_id is not NULL")),
		index.Fields("kind", "tenant_id").
			Annotations(entsql.IndexWhere("deleted_at is NULL and tenant_id is not NULL")),
		index.Fields("kind", "account_id").
			Annotations(entsql.IndexWhere("deleted_at is NULL and account_id is not NULL")),
		index.Fields("kind", "project_id").
			Annotations(entsql.IndexWhere("deleted_at is NULL and project_id is not NULL")),
		index.Fields("kind", "region").
			Annotations(entsql.IndexWhere("deleted_at is NULL and region is not NULL")),
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
