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
	"github.com/theopenlane/core/internal/ent/generated/privacy"
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
		emixin.AuditMixin{},
		emixin.IDMixin{},
		emixin.TagMixin{},
		mixin.SoftDeleteMixin{},
		NewObjectOwnedMixin(ObjectOwnedMixin{
			FieldNames:            []string{"template_id"},
			WithOrganizationOwner: true,
			Ref:                   "document_data",
		})}
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
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
		entfga.SelfAccessChecks(),
	}
}

// Interceptors of the DocumentData
func (DocumentData) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
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
