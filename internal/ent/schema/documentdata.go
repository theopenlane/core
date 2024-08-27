package schema

import (
	"context"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"

	emixin "github.com/datumforge/entx/mixin"
	"github.com/datumforge/fgax/entfga"

	"github.com/theopenlane/core/internal/ent/customtypes"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/mixin"
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
		OrgOwnerMixin{
			Ref: "documentdata",
		},
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
	}
}

// Annotations of the DocumentData
func (DocumentData) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entgql.RelayConnection(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
		entfga.Annotations{
			ObjectType:      "organization",
			IncludeHooks:    false,
			NillableIDField: true,
			OrgOwnedField:   true,
			IDField:         "OwnerID",
		},
	}
}

// Policy of the DocumentData
func (DocumentData) Policy() ent.Policy {
	return privacy.Policy{
		Mutation: privacy.MutationPolicy{
			privacy.DocumentDataMutationRuleFunc(func(ctx context.Context, m *generated.DocumentDataMutation) error {
				return m.CheckAccessForEdit(ctx)
			}),

			privacy.AlwaysDenyRule(),
		},
		Query: privacy.QueryPolicy{
			privacy.DocumentDataQueryRuleFunc(func(ctx context.Context, q *generated.DocumentDataQuery) error {
				return q.CheckAccess(ctx)
			}),
			privacy.AlwaysDenyRule(),
		},
	}
}
