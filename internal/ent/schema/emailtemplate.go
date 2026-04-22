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
	"github.com/theopenlane/iam/entfga"

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
			Optional().
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
		field.Enum("template_context").
			Comment("runtime data context defining available variable keys for this template").
			GoType(enums.TemplateContext("")).
			Optional().
			Annotations(
				entgql.OrderField("TEMPLATE_CONTEXT"),
			),
		field.JSON("defaults", map[string]any{}).
			Comment("static variable values merged as base layer at render time; call-site data takes precedence").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
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
			Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Edges of the EmailTemplate.
func (e EmailTemplate) Edges() []ent.Edge {
	return []ent.Edge{
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
		edgeToWithPagination(&edgeDefinition{
			fromSchema: e,
			edgeSchema: File{},
		}),
	}
}

// Mixin of the EmailTemplate.
func (e EmailTemplate) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags:     true,
		includeRevision: true,
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.EmailTemplate](e,
				withParents(Organization{}, Integration{}, WorkflowDefinition{}, WorkflowInstance{}),
				withOrganizationOwner(true),
			),
			mixin.NewSystemOwnedMixin(),
			newGroupPermissionsMixin(),
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
		policy.WithMutationRules(
			rule.AllowMutationIfSystemAdmin(),
			policy.CheckCreateAccess(),
			policy.CheckOrgWriteAccess(),
		),
	)
}

// Annotations of the EmailTemplate
func (EmailTemplate) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entx.FileCategory(SchemaEmailTemplate),
		entfga.SelfAccessChecks(),
	}
}
