package schema

import (
	"context"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"

	emixin "github.com/theopenlane/entx/mixin"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/mixin"
)

// Integration maps configured integrations (github, slack, etc.) to organizations
type Integration struct {
	ent.Schema
}

// Fields of the Integration
func (Integration) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name of the integration - must be unique within the organization").
			NotEmpty().
			Annotations(
				entgql.OrderField("name"),
			),
		field.String("description").
			Comment("a description of the integration").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.String("kind").
			Optional().
			Annotations(
				entgql.OrderField("kind"),
			),
	}
}

// Edges of the Integration
func (Integration) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("secrets", Hush.Type).
			Comment("the secrets associated with the integration"),
		edge.To("oauth2tokens", OhAuthTooToken.Type).
			Comment("the oauth2 tokens associated with the integration"),
		edge.To("events", Event.Type),
		edge.To("webhooks", Webhook.Type),
	}
}

// Annotations of the Integration
func (Integration) Annotations() []schema.Annotation {
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

// Mixin of the Integration
func (Integration) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		emixin.IDMixin{},
		emixin.TagMixin{},
		mixin.SoftDeleteMixin{},
		OrgOwnerMixin{
			Ref: "integrations",
		},
	}
}

// Policy of the Integration
func (Integration) Policy() ent.Policy {
	return privacy.Policy{
		Mutation: privacy.MutationPolicy{
			privacy.IntegrationMutationRuleFunc(func(ctx context.Context, m *generated.IntegrationMutation) error {
				return m.CheckAccessForDelete(ctx)
			}),

			privacy.AlwaysDenyRule(),
		},
		Query: privacy.QueryPolicy{
			privacy.IntegrationQueryRuleFunc(func(ctx context.Context, q *generated.IntegrationQuery) error {
				return q.CheckAccess(ctx)
			}),
			privacy.AlwaysDenyRule(),
		},
	}
}
