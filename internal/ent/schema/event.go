package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	emixin "github.com/theopenlane/entx/mixin"
)

// Event holds the schema definition for the Event entity
type Event struct {
	ent.Schema
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
