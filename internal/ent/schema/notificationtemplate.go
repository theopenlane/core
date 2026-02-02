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

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
)

// NotificationTemplate holds the schema definition for notification templates.
type NotificationTemplate struct {
	SchemaFuncs

	ent.Schema
}

// SchemaNotificationTemplate is the name of the NotificationTemplate schema.
const SchemaNotificationTemplate = "notification_template"

// Name returns the name of the NotificationTemplate schema.
func (NotificationTemplate) Name() string {
	return SchemaNotificationTemplate
}

// GetType returns the type of the NotificationTemplate schema.
func (NotificationTemplate) GetType() any {
	return NotificationTemplate.Type
}

// PluralName returns the plural name of the NotificationTemplate schema.
func (NotificationTemplate) PluralName() string {
	return pluralize.NewClient().Plural(SchemaNotificationTemplate)
}

// Fields of the NotificationTemplate.
func (NotificationTemplate) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").
			Comment("stable identifier for the template").
			NotEmpty().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("KEY"),
			),
		field.String("name").
			Comment("display name for the template").
			NotEmpty().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("NAME"),
			),
		field.String("description").
			Comment("description of the template").
			Optional(),
		field.Enum("channel").
			Comment("channel this template is intended for").
			GoType(enums.Channel("")).
			Annotations(
				entgql.OrderField("CHANNEL"),
			),
		field.Enum("format").
			Comment("template format for rendering").
			GoType(enums.NotificationTemplateFormat("")).
			Default(enums.NotificationTemplateFormatMarkdown.String()).
			Annotations(
				entgql.OrderField("FORMAT"),
			),
		field.String("locale").
			Comment("locale for the template, e.g. en-US").
			Default("en-US").
			Annotations(
				entgql.OrderField("LOCALE"),
			),
		field.String("topic_pattern").
			Comment("soiree topic name or wildcard pattern this template targets").
			NotEmpty().
			Annotations(
				entgql.OrderField("TOPIC_PATTERN"),
			),
		field.String("integration_id").
			Comment("integration associated with this template").
			Optional(),
		field.String("workflow_definition_id").
			Comment("workflow definition associated with this template").
			Optional(),
		field.String("email_template_id").
			Comment("optional email template used for branded email delivery").
			Optional(),
		field.String("title_template").
			Comment("title template for external channel messages").
			Optional(),
		field.String("subject_template").
			Comment("subject template for email notifications").
			Optional(),
		field.Text("body_template").
			Comment("body template for the notification").
			Optional(),
		field.JSON("blocks", map[string]any{}).
			Comment("structured blocks for channels like Slack or Teams").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.JSON("jsonconfig", map[string]any{}).
			Comment("jsonschema for template data requirements").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.JSON("uischema", map[string]any{}).
			Comment("uischema for a template builder").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.JSON("metadata", map[string]any{}).
			Comment("additional template metadata").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.Bool("active").
			Comment("whether the template is active").
			Default(true).
			Annotations(
				entgql.OrderField("ACTIVE"),
			),
		field.Int("version").
			Comment("template version").
			Default(1).
			Annotations(
				entgql.OrderField("VERSION"),
			),
	}
}

// Indexes of the NotificationTemplate.
func (NotificationTemplate) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields(ownerFieldName, "channel", "locale", "topic_pattern").
			Unique().
			Annotations(entsql.IndexWhere("deleted_at is NULL")),
		index.Fields(ownerFieldName, "key").
			Unique().
			Annotations(entsql.IndexWhere("deleted_at is NULL")),
		index.Fields("key").
			Unique().
			Annotations(entsql.IndexWhere("deleted_at is NULL and system_owned = true")),
	}
}

// Edges of the NotificationTemplate.
func (n NotificationTemplate) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: n,
			edgeSchema: Integration{},
			field:      "integration_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Integration{}.Name()),
			},
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: n,
			edgeSchema: WorkflowDefinition{},
			field:      "workflow_definition_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(WorkflowDefinition{}.Name()),
			},
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: n,
			edgeSchema: EmailTemplate{},
			field:      "email_template_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(EmailTemplate{}.Name()),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: n,
			edgeSchema: Notification{},
		}),
	}
}

// Mixin of the NotificationTemplate.
func (n NotificationTemplate) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags: true,
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.NotificationTemplate](n,
				withOrganizationOwner(true),
			),
			mixin.NewSystemOwnedMixin(),
		},
	}.getMixins(n)
}

// Modules of the NotificationTemplate.
func (NotificationTemplate) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogBaseModule,
	}
}

// Policy of the NotificationTemplate.
func (NotificationTemplate) Policy() ent.Policy {
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
