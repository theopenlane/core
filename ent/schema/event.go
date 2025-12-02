package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx/history"
	emixin "github.com/theopenlane/entx/mixin"

	"github.com/theopenlane/ent/generated/privacy"
	"github.com/theopenlane/ent/mixin"
	"github.com/theopenlane/ent/privacy/policy"
	"github.com/theopenlane/shared/models"
)

// Event holds the schema definition for the Event entity
type Event struct {
	SchemaFuncs

	ent.Schema
}

// SchemaEvent is the name of the Event schema.
const SchemaEvent = "event"

// Name returns the name of the Event schema.
func (Event) Name() string {
	return SchemaEvent
}

// GetType returns the type of the Event schema.
func (Event) GetType() any {
	return Event.Type
}

// PluralName returns the plural name of the Event schema.
func (Event) PluralName() string {
	return pluralize.NewClient().Plural(SchemaEvent)
}

// Fields of the Event
func (Event) Fields() []ent.Field {
	return []ent.Field{
		field.String("event_id").
			Optional(),
		field.String("correlation_id").
			Optional(),
		field.String("event_type"),
		field.JSON("metadata", map[string]any{}).Optional(),
	}
}

// Edges of the Event
func (e Event) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeFromWithPagination(e, User{}),
		defaultEdgeFromWithPagination(e, Group{}),
		defaultEdgeFromWithPagination(e, Integration{}),
		defaultEdgeFromWithPagination(e, Organization{}),
		defaultEdgeFromWithPagination(e, Invite{}),
		defaultEdgeFromWithPagination(e, PersonalAccessToken{}),
		defaultEdgeFromWithPagination(e, Hush{}),
		defaultEdgeFromWithPagination(e, OrgMembership{}),
		defaultEdgeFromWithPagination(e, GroupMembership{}),
		defaultEdgeFromWithPagination(e, Subscriber{}),
		defaultEdgeFromWithPagination(e, File{}),
		defaultEdgeFromWithPagination(e, OrgSubscription{}),
	}
}

func (Event) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogBaseModule,
	}
}

// Annotations of the Event
func (e Event) Annotations() []schema.Annotation {
	return []schema.Annotation{
		history.Annotations{
			Exclude: true,
		},
	}
}

// Mixin of the Event
func (Event) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		emixin.IDMixin{},
		mixin.TagMixin{},
		mixin.GraphQLAnnotationMixin{},
	}
}

// Policy of the Event
func (e Event) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			// allow after interceptors are properly added
			privacy.AlwaysDenyRule(),
		),
		policy.WithMutationRules(),
	)
}
