package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/theopenlane/entx"
	emixin "github.com/theopenlane/entx/mixin"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/pkg/enums"
)

// Template holds the schema definition for the Template entity
type Template struct {
	ent.Schema
}

// Mixin of the Template
func (Template) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		mixin.SoftDeleteMixin{},
		emixin.IDMixin{},
		emixin.TagMixin{},
		NewOrgOwnMixinWithRef("templates"),
	}
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
			Default(string(enums.Document)),
		field.String("description").
			Comment("the description of the template").
			Optional(),
		field.JSON("jsonconfig", map[string]any{}).
			Comment("the jsonschema object of the template").
			Annotations(
				entx.FieldJSONPathSearchable("$id"),
			),
		field.JSON("uischema", map[string]any{}).
			Comment("the uischema for the template to render in the UI").
			Optional(),
	}
}

// Edges of the Template
func (Template) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("documents", DocumentData.Type).
			Annotations(
				entx.CascadeAnnotationField("Template"),
			),
		edge.To("files", File.Type),
	}
}

// Indexes of the Template
func (Template) Indexes() []ent.Index {
	return []ent.Index{
		// names should be unique, but ignore deleted names
		index.Fields("name", ownerFieldName, "template_type").
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Annotations of the Template
func (Template) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
		entfga.OrganizationInheritedChecks(),
	}
}

// Policy of the Template
func (Template) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(), //  interceptor should filter out the results
		),
		policy.WithMutationRules(
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.TemplateMutation](),
		),
	)
}
