package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/hooks"
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
		field.String("platform_id").
			Comment("optional platform associated with this integration for downstream inventory linkage").
			Optional().
			NotEmpty().
			Immutable(),
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
		field.JSON("installation_metadata", openapi.IntegrationInstallationMetadata{}).
			Comment("stable, non-secret installation identity metadata for the provider").
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
		field.String("definition_id").
			Comment("the canonical definition identifier for the installation").
			Optional().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("definition_id"),
			),
		field.String("definition_version").
			Comment("the definition version recorded for this installation").
			Optional().
			Annotations(
				entgql.OrderField("definition_version"),
			),
		field.String("definition_slug").
			Comment("the human-readable definition slug recorded for this installation").
			Optional().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("definition_slug"),
			),
		field.String("family").
			Comment("the denormalized family label for the installation definition").
			Optional().
			Annotations(
				entgql.OrderField("family"),
			),
		field.Enum("status").
			Comment("the lifecycle status of the installation").
			GoType(enums.IntegrationStatus("")).
			Default(enums.IntegrationStatusPending.String()).
			Annotations(
				entgql.OrderField("status"),
			),
		field.JSON("provider_metadata_snapshot", map[string]any{}).
			Comment("snapshot of definition metadata captured on the installation").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.Bool("primary_directory").
			Comment("designates this integration as the authoritative directory source for identity holder enrichment and lifecycle derivation within its owner organization").
			Default(false),
		field.Bool("campaign_email").
			Comment("designates this email integration as the one to use for campaign dispatch within its owner organization").
			Default(false),
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
		defaultEdgeToWithPagination(i, Asset{}),
		defaultEdgeToWithPagination(i, DirectoryAccount{}),
		defaultEdgeToWithPagination(i, DirectoryGroup{}),
		defaultEdgeToWithPagination(i, DirectoryMembership{}),
		defaultEdgeToWithPagination(i, DirectorySyncRun{}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: i,
			edgeSchema: Platform{},
			field:      "platform_id",
			immutable:  true,
			comment:    "platform associated with this integration",
		}),
		defaultEdgeToWithPagination(i, NotificationTemplate{}),
		defaultEdgeToWithPagination(i, EmailTemplate{}),
		defaultEdgeToWithPagination(i, Campaign{}),
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

// Hooks of the Integration
func (Integration) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookIntegrationPrimaryDirectory(),
		hooks.HookIntegrationCampaignEmail(),
	}
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
