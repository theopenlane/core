package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/hooks"
	"github.com/theopenlane/ent/privacy/policy"
	"github.com/theopenlane/shared/models"
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
			Optional().
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
		}),
		defaultEdgeFromWithPagination(d, Entity{}),
		defaultEdgeToWithPagination(d, File{}),
	}
}

func (DocumentData) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
	}
}

// Annotations of the DocumentData
func (d DocumentData) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
	}
}

// Policy of the DocumentData
func (d DocumentData) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			// TODO: this should ensure the correct access for creation
			// it currently only checks edit access
			entfga.CheckEditAccess[*generated.DocumentDataMutation](),
		),
	)
}

func (d DocumentData) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookDocumentDataTrustCenterNDA(),
	}
}
