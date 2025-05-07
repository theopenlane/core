package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/pkg/enums"
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

// Mixin of the Template
func (t Template) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(t),
		},
	}.getMixins()
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

// Policy of the Template
func (Template) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(), //  interceptor should filter out the results
		),
		policy.WithMutationRules(
			policy.CheckCreateAccess(),
			policy.CheckOrgWriteAccess()),
	)
}
