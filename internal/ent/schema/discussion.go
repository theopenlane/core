package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/iam/entfga"
)

// Discussion holds the schema definition for the Discussion entity
type Discussion struct {
	SchemaFuncs

	ent.Schema
}

// SchemaDiscussionis the name of the schema in snake case
const SchemaDiscussion = "discussion"

// Name is the name of the schema in snake case
func (Discussion) Name() string {
	return SchemaDiscussion
}

// GetType returns the type of the schema
func (Discussion) GetType() any {
	return Discussion.Type
}

// PluralName returns the plural name of the schema
func (Discussion) PluralName() string {
	return pluralize.NewClient().Plural(SchemaDiscussion)
}

// Fields of the Discussion
func (Discussion) Fields() []ent.Field {
	return []ent.Field{
		field.String("external_id").
			Unique().
			Comment("the unique discussion identifier from external system, e.g. plate discussion id"),
		field.Bool("is_resolved").
			Comment("whether the discussion is resolved").
			Default(false),
	}
}

// Mixin of the Discussion
func (d Discussion) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags: true,
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(d),
		},
	}.getMixins(d)
}

// Edges of the Discussion
func (d Discussion) Edges() []ent.Edge {
	return []ent.Edge{
		edgeToWithPagination(&edgeDefinition{
			fromSchema: d,
			t:          Note.Type,
			name:       "comments",
			comment:    "the comments in the discussion",
			annotations: []schema.Annotation{
				accessmap.EdgeNoAuthCheck(),
			},
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: d,
			edgeSchema: Control{},
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Control{}.Name()),
			},
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: d,
			edgeSchema: Subcontrol{},
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Subcontrol{}.Name()),
			},
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: d,
			edgeSchema: Procedure{},
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Procedure{}.Name()),
			},
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: d,
			edgeSchema: Risk{},
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Risk{}.Name()),
			},
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: d,
			edgeSchema: InternalPolicy{},
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(InternalPolicy{}.Name()),
			},
		}),
	}
}

// Indexes of the Discussion
func (Discussion) Indexes() []ent.Index {
	return []ent.Index{}
}

// Annotations of the Discussion
func (Discussion) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
	}
}

// Hooks of the Discussion
func (Discussion) Hooks() []ent.Hook {
	return []ent.Hook{}
}

// Interceptors of the Discussion
func (Discussion) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}

// Modules this schema has access to
func (Discussion) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogBaseModule,
	}
}

// Policy of the Discussion
func (Discussion) Policy() ent.Policy {
	// add the new policy here, the default post-policy is to deny all
	// so you need to ensure there are rules in place to allow the actions you want
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.AllowCreate(),
			entfga.CheckEditAccess[*generated.DiscussionMutation](),
		),
	)
}
