package schema

import (
	"time"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"

	"github.com/datumforge/enthistory"
	emixin "github.com/datumforge/entx/mixin"
)

// OhAuthTooToken holds the schema definition for the OhAuthTooToken entity
type OhAuthTooToken struct {
	ent.Schema
}

// Fields of the OhAuthTooToken
func (OhAuthTooToken) Fields() []ent.Field {
	return []ent.Field{
		field.Text("client_id").
			NotEmpty(),
		field.JSON("scopes", []string{}).
			Optional(),
		field.Text("nonce").
			NotEmpty(),
		field.Text("claims_user_id").
			NotEmpty(),
		field.Text("claims_username").
			NotEmpty(),
		field.Text("claims_email").
			NotEmpty(),
		field.Bool("claims_email_verified"),
		field.JSON("claims_groups", []string{}).
			Optional(),
		field.Text("claims_preferred_username"),
		field.Text("connector_id").
			NotEmpty(),
		field.JSON("connector_data", []string{}).
			Optional(),
		field.Time("last_used").
			Default(time.Now),
	}
}

// Edges of the OhAuthTooToken
func (OhAuthTooToken) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("integration", Integration.Type).
			Ref("oauth2tokens"),
		edge.To("events", Event.Type),
	}
}

// Mixin of the OhAuthTooToken
func (OhAuthTooToken) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.IDMixin{},
		emixin.TagMixin{},
	}
}

// Annotations of the OhAuthTooToken
func (OhAuthTooToken) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entgql.RelayConnection(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
		enthistory.Annotations{
			Exclude: true,
		},
	}
}
