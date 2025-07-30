package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
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
	}.getMixins(d)
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

func (DocumentData) Features() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogBaseModule,
	}
}

// Annotations of the DocumentData
func (d DocumentData) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
	}
}

// Interceptors of the DocumentData
func (d DocumentData) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorRequireAnyFeature("documentdata", d.Features()...),
	}
}

// Policy of the DocumentData
func (d DocumentData) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			rule.DenyIfMissingAllFeatures(d.Features()...),
			entfga.CheckEditAccess[*generated.DocumentDataMutation](),
		),
	)
}
