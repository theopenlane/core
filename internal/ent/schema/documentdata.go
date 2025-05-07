package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
)

// DocumentData holds the schema definition for the DocumentData entity
type DocumentData struct {
	SchemaFuncs

	ent.Schema
}

// SchemaDocumentData is the name of the DocumentData schema.
const SchemaDocumentData = "document"

// Name returns the name of the DocumentData schema.
func (DocumentData) Name() string {
	return SchemaDocumentData
}

// GetType returns the type of the DocumentData schema.
func (DocumentData) GetType() any {
	return DocumentData.Type
}

// PluralName returns the plural name of the DocumentData schema.
func (DocumentData) PluralName() string {
	return pluralize.NewClient().Plural(SchemaDocumentData)
}

// Fields of the DocumentData
func (DocumentData) Fields() []ent.Field {
	return []ent.Field{
		field.String("template_id").
			Comment("the template id of the document"),
		field.JSON("data", map[string]any{}).
			Comment("the json data of the document"),
	}
}

// Mixin of the DocumentData
func (d DocumentData) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.DocumentData](d,
				withOrganizationOwner(false),
				withParents(Template{})),
		},
	}.getMixins()
}

// Edges of the DocumentData
func (d DocumentData) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: d,
			edgeSchema: Template{},
			field:      "template_id",
			required:   true,
		}),
		defaultEdgeFromWithPagination(d, Entity{}),
		defaultEdgeToWithPagination(d, File{}),
	}
}

// Annotations of the DocumentData
func (DocumentData) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
	}
}

// Policy of the DocumentData
func (DocumentData) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(), //  interceptor should filter out the results
		),
		policy.WithMutationRules(
			entfga.CheckEditAccess[*generated.DocumentDataMutation](),
		),
	)
}
