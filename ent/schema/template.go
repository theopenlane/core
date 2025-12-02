package schema

import (
	"fmt"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/hooks"
	"github.com/theopenlane/ent/mixin"
	"github.com/theopenlane/ent/privacy/policy"
	"github.com/theopenlane/ent/privacy/rule"
	"github.com/theopenlane/shared/enums"
)

// Template holds the schema definition for the Template entity
type Template struct {
	SchemaFuncs

	ent.Schema
}

// SchemaTemplate is the name of the Template schema.
const SchemaTemplate = "template"

// SchemaTemplate is the name of the Template schema.
func (Template) Name() string {
	return SchemaTemplate
}

// GetType returns the type of the Template schema.
func (Template) GetType() any {
	return Template.Type
}

// PluralName returns the plural name of the Template schema.
func (Template) PluralName() string {
	return pluralize.NewClient().Plural(SchemaTemplate)
}

// Fields of the Template
func (Template) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name of the template").
			NotEmpty().
			Annotations(
				entgql.OrderField("name"),
				entx.FieldSearchable(),
			),
		field.Enum("template_type").
			Comment("the type of the template, either a provided template or an implementation (document)").
			GoType(enums.DocumentType("")).
			Annotations(
				entgql.OrderField("TEMPLATE_TYPE"),
			).
			Default(enums.Document.String()),

		field.String("description").
			Comment("the description of the template").
			Optional(),

		field.Enum("kind").
			Comment("the kind of template, e.g. questionnaire").
			Optional().
			GoType(enums.TemplateKind("")).
			Default(enums.TemplateKindQuestionnaire.String()).
			Annotations(
				entgql.OrderField("KIND"),
			),

		field.JSON("jsonconfig", map[string]any{}).
			Comment("the jsonschema object of the template").
			Annotations(
				entx.FieldJSONPathSearchable("$id"),
			),
		field.JSON("uischema", map[string]any{}).
			Comment("the uischema for the template to render in the UI").
			Optional(),
		field.String("trust_center_id").
			Comment("the id of the trust center this template is associated with").
			Optional(),
	}
}

// Mixin of the Template
func (t Template) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.Template](t,
				withParents(Organization{}, TrustCenter{}),
				withOrganizationOwner(true),
				withSkipperFunc(skipInterceptorForOrgMembers),
				withAllowAnonymousTrustCenterAccess(true),
			),
			mixin.NewSystemOwnedMixin(),
		},
	}.getMixins(t)
}

// Edges of the Template
func (t Template) Edges() []ent.Edge {
	return []ent.Edge{
		edgeToWithPagination(&edgeDefinition{
			fromSchema:    t,
			edgeSchema:    DocumentData{},
			cascadeDelete: "Template",
		}),
		defaultEdgeToWithPagination(t, File{}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: t,
			edgeSchema: TrustCenter{},
			field:      "trust_center_id",
		}),

		edgeToWithPagination(&edgeDefinition{
			fromSchema: t,
			edgeSchema: Assessment{},
		}),
	}
}

// Indexes of the Template
func (Template) Indexes() []ent.Index {
	return []ent.Index{
		// names should be unique, but ignore deleted names
		index.Fields("name", ownerFieldName, "template_type").
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
		// Only one non-deleted NDA per trust center allowed
		index.Fields("trust_center_id").
			Unique().Annotations(entsql.IndexWhere(fmt.Sprintf("deleted_at is NULL and kind = '%s'", enums.TemplateKindTrustCenterNda.String()))),
	}
}

func (Template) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Check: fmt.Sprintf("trust_center_id IS NOT NULL OR kind != '%s'", enums.TemplateKindTrustCenterNda.String()),
		},
		entfga.SelfAccessChecks(),
	}
}

// Hooks of the Template
func (Template) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookTemplate(),
		hooks.HookTemplateFiles(),
		hooks.HookTemplateAuthz(),
	}
}

// Policy of the Template
func (Template) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			rule.AllowMutationIfSystemAdmin(),
			policy.CheckCreateAccess(),
			policy.CheckOrgWriteAccess()),
	)
}
