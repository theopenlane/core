package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	emixin "github.com/theopenlane/entx/mixin"
)

type Event struct {
	ent.Schema
}

// Fields of the TicketEvent.
func (Event) Fields() []ent.Field {
	return []ent.Field{
		field.String("event_id").
			Comment("the unique identifier of the event as it relates to the source or outside system").
			Unique().
			Immutable(),
		field.String("correlation_id").
			Comment("an identifier to correleate the event to another object or source, if needed").
			Optional(),
		field.String("event_type").
			Optional().
			Comment("the type of event"),
		field.JSON("metadata", map[string]any{}).
			Comment("event metadata").
			Optional(),
		field.String("source").
			Comment("the source of the event").
			Optional(),
		field.Bool("additional_processing_required").
			Default(false).
			Comment("indicates if additional processing is required for the event").
			Optional(),
		field.String("additional_processing_details").
			Comment("details about the additional processing required").
			Optional(),
		field.String("processed_by").
			Comment("the listener ID who processed the event").
			Optional(),
		field.Time("processed_at").
			Comment("the time the event was processed").
			Optional(),
	}
}

// Edges of the Event
func (Event) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("events"),
		edge.From("group", Group.Type).Ref("events"),
		edge.From("integration", Integration.Type).Ref("events"),
		edge.From("organization", Organization.Type).Ref("events"),
		edge.From("invite", Invite.Type).Ref("events"),
		edge.From("personal_access_token", PersonalAccessToken.Type).Ref("events"),
		edge.From("hush", Hush.Type).Ref("events"),
		edge.From("orgmembership", OrgMembership.Type).Ref("events"),
		edge.From("groupmembership", GroupMembership.Type).Ref("events"),
		edge.From("subscriber", Subscriber.Type).Ref("events"),
		edge.From("file", File.Type).Ref("events"),
		edge.From("orgsubscription", OrgSubscription.Type).Ref("events"),
	}
}

// Annotations of the Event
func (Event) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entgql.RelayConnection(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
	}
}

// Mixin of the Event
func (Event) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		emixin.IDMixin{},
		emixin.TagMixin{},
	}
}
