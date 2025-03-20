package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/core/internal/ent/mixin"
	emixin "github.com/theopenlane/entx/mixin"
)

// Event holds the schema definition for the Event entity
type Event struct {
	CustomSchema

	ent.Schema
}

const SchemaEvent = "event"

func (Event) Name() string {
	return SchemaEvent
}

func (Event) GetType() any {
	return Event.Type
}

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

// Annotations of the Event
func (Event) Annotations() []schema.Annotation {
	return []schema.Annotation{}
}

// Mixin of the Event
func (Event) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		emixin.IDMixin{},
		emixin.TagMixin{},
		mixin.AnnotationMixin{},
	}
}
