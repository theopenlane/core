package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/entx/history"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/validator"
	"github.com/theopenlane/utils/keygen"
)

// IntegrationWebhook holds the schema definition for integration webhooks.
type IntegrationWebhook struct {
	SchemaFuncs

	ent.Schema
}

// SchemaIntegrationWebhook is the name of the IntegrationWebhook schema.
const SchemaIntegrationWebhook = "integration_webhook"

// Name returns the name of the IntegrationWebhook schema.
func (IntegrationWebhook) Name() string {
	return SchemaIntegrationWebhook
}

// GetType returns the type of the IntegrationWebhook schema.
func (IntegrationWebhook) GetType() any {
	return IntegrationWebhook.Type
}

// PluralName returns the plural name of the IntegrationWebhook schema.
func (IntegrationWebhook) PluralName() string {
	return pluralize.NewClient().Plural(SchemaIntegrationWebhook)
}

// Fields of the IntegrationWebhook.
func (IntegrationWebhook) Fields() []ent.Field {
	return []ent.Field{
		field.String("integration_id").
			Comment("integration connection this webhook belongs to").
			Optional(),
		field.String("name").
			Comment("display name for the webhook endpoint").
			Optional().
			Annotations(
				entgql.OrderField("NAME"),
			),
		field.String("status").
			Comment("status of the webhook endpoint").
			Optional().
			Annotations(
				entgql.OrderField("STATUS"),
			),
		field.String("endpoint_url").
			Comment("destination URL for webhook delivery").
			Optional().
			Validate(validator.ValidateURL()).
			Nillable().
			Annotations(
				entgql.OrderField("ENDPOINT_URL"),
			),
		field.String("secret_token").
			Comment("secret token for webhook signature validation").
			Sensitive().
			Immutable().
			Optional().
			DefaultFunc(func() string {
				return keygen.PrefixedSecret("tola_wsec")
			}).
			Annotations(
				entgql.Skip(entgql.SkipWhereInput, entgql.SkipMutationUpdateInput),
			),
		field.Strings("allowed_events").
			Comment("list of allowed event types for this webhook").
			Optional(),
		field.String("last_delivery_id").
			Comment("identifier of the last delivery attempt").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipMutationUpdateInput),
			),
		field.Time("last_delivery_at").
			Comment("timestamp of the last delivery attempt").
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("LAST_DELIVERY_AT"),
				entgql.Skip(entgql.SkipMutationUpdateInput),
			),
		field.String("last_delivery_status").
			Comment("status of the last delivery attempt").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipMutationUpdateInput),
			),
		field.Text("last_delivery_error").
			Comment("error details from the last delivery attempt").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipMutationUpdateInput),
			),
		field.JSON("metadata", map[string]any{}).
			Comment("additional webhook metadata").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
	}
}

// Edges of the IntegrationWebhook.
func (w IntegrationWebhook) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: w,
			edgeSchema: Integration{},
			field:      "integration_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Integration{}.Name()),
			},
		}),
	}
}

// Mixin of the IntegrationWebhook.
func (w IntegrationWebhook) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags: true,
		excludeAnnotations: true,
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[IntegrationWebhook](w,
				withOrganizationOwner(true),
			),
		},
	}.getMixins(w)
}

// Modules of the IntegrationWebhook.
func (IntegrationWebhook) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogBaseModule,
	}
}

// Policy of the IntegrationWebhook.
//func (IntegrationWebhook) Policy() ent.Policy {
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

// Annotations of the IntegrationWebhook.
func (r IntegrationWebhook) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.Skip(entgql.SkipAll),
		entx.SchemaGenSkip(true),
		entx.QueryGenSkip(true),
		history.Annotations{
			Exclude: true,
		},
	}
}
