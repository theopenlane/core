package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"

	emixin "github.com/theopenlane/entx/mixin"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/customtypes"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
)

// DocumentData holds the schema definition for the DocumentData entity
type DocumentData struct {
	ent.Schema
}

// Fields of the DocumentData
func (DocumentData) Fields() []ent.Field {
	return []ent.Field{
		field.String("template_id").
			Comment("the template id of the document"),
		field.JSON("data", customtypes.JSONObject{}).
			Comment("the json data of the document").
			Annotations(
				entgql.Type("JSON"),
			),
	}
}

// Mixin of the DocumentData
func (DocumentData) Mixin() []ent.Mixin {
	return []ent.Mixin{
		NewAuditMixin(),
		emixin.IDMixin{},
		emixin.TagMixin{},
		mixin.SoftDeleteMixin{},
		NewOrgOwnMixinWithRef("document_data"),
	}
}

// Edges of the DocumentData
func (DocumentData) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("template", Template.Type).
			Ref("documents").
			Unique().
			Required().
			Field("template_id"),
		edge.From("entity", Entity.Type).
			Ref("documents"),
		edge.To("files", File.Type),
	}
}

// Annotations of the DocumentData
func (DocumentData) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entgql.RelayConnection(),
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entfga.OrganizationInheritedChecks(), // TODO(sfunk): update to template checks instead of org checks
	}
}

// Policy of the DocumentData
func (DocumentData) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			entfga.CheckReadAccess[*generated.DocumentDataQuery](),
		),
		policy.WithMutationRules(
			entfga.CheckEditAccess[*generated.DocumentDataMutation](),
		),
	)
}
