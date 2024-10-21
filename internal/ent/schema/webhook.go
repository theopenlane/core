package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/theopenlane/entx/history"
	emixin "github.com/theopenlane/entx/mixin"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/mixin"
)

// Webhook holds the schema definition for the Webhook entity
type Webhook struct {
	ent.Schema
}

// Fields of the Webhook
func (Webhook) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name of the webhook").
			NotEmpty().
			Annotations(
				entgql.OrderField("name"),
			),
		field.String("description").
			Comment("a description of the webhook").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.String("destination_url").
			Comment("the url to send the webhook to").
			NotEmpty().
			Annotations(
				entgql.OrderField("url"),
			),
		field.Bool("enabled").
			Comment("indicates if the webhook is active and enabled").
			Default(true),
		field.String("callback").
			Comment("the call back string").
			Unique().
			Annotations(entgql.Skip()).
			Optional(),
		field.Time("expires_at").
			Comment("the ttl of the webhook delivery").
			Annotations(entgql.Skip()).
			Optional(),
		field.Bytes("secret").
			Comment("the comparison secret to verify the token's signature").
			Annotations(entgql.Skip()).
			Optional(),
		field.Int("failures").
			Comment("the number of failures").
			Optional().
			Default(0),
		field.String("last_error").
			Comment("the last error message").
			Optional(),
		field.String("last_response").
			Comment("the last response").
			Optional(),
	}
}

// Mixin of the Webhook
func (Webhook) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		emixin.IDMixin{},
		emixin.TagMixin{},
		mixin.SoftDeleteMixin{},
		NewOrgOwnedMixin(ObjectOwnedMixin{
			Ref:        "webhooks",
			AllowEmpty: true,
		}),
	}
}

// Edges of the Webhook
func (Webhook) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("events", Event.Type),
		edge.From("integrations", Integration.Type).Ref("webhooks"),
	}
}

// Indexes of the Webhook
func (Webhook) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name", "owner_id").
			Unique().
			Annotations(
				entsql.IndexWhere("deleted_at is NULL"),
			),
	}
}

// Annotations of the Webhook
func (Webhook) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entgql.RelayConnection(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
		entfga.Annotations{
			ObjectType:      "organization",
			IncludeHooks:    false,
			IDField:         "OwnerID",
			NillableIDField: true,
		},
		history.Annotations{
			Exclude: false,
		},
	}
}
