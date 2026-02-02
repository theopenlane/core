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

// EmailTemplate holds the schema definition for email templates.
type EmailTemplate struct {
	SchemaFuncs

	ent.Schema
}

// SchemaEmailTemplate is the name of the EmailTemplate schema.
const SchemaEmailTemplate = "email_template"

// Name returns the name of the EmailTemplate schema.
func (EmailTemplate) Name() string {
	return SchemaEmailTemplate
}

// GetType returns the type of the EmailTemplate schema.
func (EmailTemplate) GetType() any {
	return EmailTemplate.Type
}

// PluralName returns the plural name of the EmailTemplate schema.
func (EmailTemplate) PluralName() string {
	return pluralize.NewClient().Plural(SchemaEmailTemplate)
}

// Fields of the EmailTemplate.
func (EmailTemplate) Fields() []ent.Field {
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
		field.Enum("format").
			Comment("template format for rendering").
			GoType(enums.NotificationTemplateFormat("")).
			Default(enums.NotificationTemplateFormatHTML.String()).
			Annotations(
				entgql.OrderField("FORMAT"),
			),
		field.String("locale").
			Comment("locale for the template, e.g. en-US").
			Default("en-US").
			Annotations(
				entgql.OrderField("LOCALE"),
			),
		field.String("subject_template").
			Comment("subject template for email notifications").
			Optional(),
		field.String("preheader_template").
			Comment("preheader/preview text template for email notifications").
			Optional(),
		field.Text("body_template").
			Comment("body template for the email").
			Optional(),
		field.Text("text_template").
			Comment("plain text fallback template for the email").
			Optional(),
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
		field.String("email_branding_id").
			Comment("email branding configuration to apply for this template").
			Optional(),
		field.String("integration_id").
			Comment("integration used to deliver emails for this template").
			Optional(),
		field.String("workflow_definition_id").
			Comment("workflow definition associated with this template").
			Optional(),
		field.String("workflow_instance_id").
			Comment("workflow instance associated with this template").
			Optional(),
	}
}

// Indexes of the EmailTemplate.
func (EmailTemplate) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields(ownerFieldName, "key").
			Unique().
			Annotations(entsql.IndexWhere("deleted_at is NULL")),
		index.Fields("key").
			Unique().
			Annotations(entsql.IndexWhere("deleted_at is NULL and system_owned = true")),
	}
}

// Edges of the EmailTemplate.
func (e EmailTemplate) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: e,
			edgeSchema: EmailBranding{},
			field:      "email_branding_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(EmailBranding{}.Name()),
			},
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: e,
			edgeSchema: Integration{},
			field:      "integration_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Integration{}.Name()),
			},
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: e,
			edgeSchema: WorkflowDefinition{},
			field:      "workflow_definition_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(WorkflowDefinition{}.Name()),
			},
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: e,
			edgeSchema: WorkflowInstance{},
			field:      "workflow_instance_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(WorkflowInstance{}.Name()),
			},
		}),
		defaultEdgeToWithPagination(e, Campaign{}),
		defaultEdgeToWithPagination(e, NotificationTemplate{}),
	}
}

// Mixin of the EmailTemplate.
func (e EmailTemplate) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags: true,
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.EmailTemplate](e,
				withOrganizationOwner(true),
			),
			mixin.NewSystemOwnedMixin(),
		},
	}.getMixins(e)
}

// Modules of the EmailTemplate.
func (EmailTemplate) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogBaseModule,
	}
}

// Policy of the EmailTemplate.
func (EmailTemplate) Policy() ent.Policy {
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
