package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/pkg/models"
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

// plate_discussion
// id (number) - ULID (not a number)
// document_id (string) - external_id **
// is_resolved (boolean) default false is_resolved**
// user_id (string) - created_by
// comment_ids (json_array)  array of id-s or many to many table (edge to comments)
// created_at (DateTime)
// Fields of the Discussion
func (Discussion) Fields() []ent.Field {
	return []ent.Field{
		field.Int("external_id").
			Unique().
			Comment("the unique discussion identifier from external system, e.g. plate discussion id"),
		field.Bool("is_resolved").
			Comment("whether the discussion is resolved").
			Default(false),
		field.String("internal_policy_id").
			Comment("the internal policy associated with the discussion").
			Optional(),
		field.String("procedure_id").
			Comment("the procedure associated with the discussion").
			Optional(),
		field.String("risk_id").
			Comment("the risk associated with the discussion").
			Optional(),
		field.String("control_id").
			Comment("the control associated with the discussion").
			Optional(),
		field.String("subcontrol_id").
			Comment("the subcontrol associated with the discussion").
			Optional(),
	}
}

// Mixin of the Discussion
func (d Discussion) Mixin() []ent.Mixin {
	// getDefaultMixins returns the default mixins for all entities
	// see mixingConfig{} for more configuration options
	return getDefaultMixins(d)
}

// Edges of the Discussion
func (d Discussion) Edges() []ent.Edge {
	return []ent.Edge{
		edgeToWithPagination(&edgeDefinition{
			fromSchema: d,
			t:          Note.Type,
			name:       "comments",
			comment:    "the comments in the discussion",
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: d,
			edgeSchema: InternalPolicy{},
			field:      "internal_policy_id",
			comment:    "the internal policy associated with the discussion",
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: d,
			edgeSchema: Procedure{},
			field:      "procedure_id",
			comment:    "the procedure associated with the discussion",
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: d,
			edgeSchema: Risk{},
			field:      "risk_id",
			comment:    "the risk associated with the discussion",
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: d,
			edgeSchema: Control{},
			field:      "control_id",
			comment:    "the control associated with the discussion",
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: d,
			edgeSchema: Subcontrol{},
			field:      "subcontrol_id",
			comment:    "the subcontrol associated with the discussion",
		}),
	}
}

// Indexes of the Discussion
func (Discussion) Indexes() []ent.Index {
	return []ent.Index{}
}

// Annotations of the Discussion
func (Discussion) Annotations() []schema.Annotation {
	return []schema.Annotation{}
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
			// add mutation rules here, the below is the recommended default
			policy.CheckCreateAccess(),
			// this needs to be commented out for the first run that had the entfga annotation
			// the first run will generate the functions required based on the entfa annotation
			// entfga.CheckEditAccess[*generated.DiscussionMutation](),
		),
	)
}
