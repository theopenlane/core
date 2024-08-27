package schema

import (
	"context"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/datumforge/entx"
	emixin "github.com/datumforge/entx/mixin"
	"github.com/datumforge/fgax/entfga"

	"github.com/theopenlane/core/internal/ent/customtypes"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/mixin"
)

// OauthProvider holds the schema definition for the OauthProvider entity
type OauthProvider struct {
	ent.Schema
}

// Fields of the OauthProvider
func (OauthProvider) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the oauth provider's name"),
		field.String("client_id").
			Comment("the client id for the oauth provider"),
		field.String("client_secret").
			Comment("the client secret"),
		field.String("redirect_url").
			Comment("the redirect url"),
		field.String("scopes").
			Comment("the scopes"),
		field.String("auth_url").
			Comment("the auth url of the provider"),
		field.String("token_url").
			Comment("the token url of the provider"),
		field.Uint8("auth_style").
			GoType(customtypes.Uint8(0)).
			Annotations(
				entgql.Type("Uint"),
			).
			Comment("the auth style, 0: auto detect 1: third party log in 2: log in with username and password"),
		field.String("info_url").
			Comment("the URL to request user information by token"),
	}
}

// Edges of the OauthProvider
func (OauthProvider) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Annotations of the OauthProvider
func (OauthProvider) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entgql.RelayConnection(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
		entx.QueryGenSkip(true),
		entfga.Annotations{
			ObjectType:      "organization",
			IncludeHooks:    false,
			NillableIDField: true,
			OrgOwnedField:   true,
			IDField:         "OwnerID",
		},
	}
}

// Mixin of the OauthProvider
func (OauthProvider) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		emixin.IDMixin{},
		emixin.TagMixin{},
		mixin.SoftDeleteMixin{},
		OrgOwnerMixin{
			Ref: "oauthprovider",
		},
	}
}

// Policy of the OauthProvider
func (OauthProvider) Policy() ent.Policy {
	return privacy.Policy{
		Mutation: privacy.MutationPolicy{
			privacy.OauthProviderMutationRuleFunc(func(ctx context.Context, m *generated.OauthProviderMutation) error {
				return m.CheckAccessForEdit(ctx)
			}),

			privacy.AlwaysDenyRule(),
		},
		Query: privacy.QueryPolicy{
			privacy.OauthProviderQueryRuleFunc(func(ctx context.Context, q *generated.OauthProviderQuery) error {
				return q.CheckAccess(ctx)
			}),
			privacy.AlwaysDenyRule(),
		},
	}
}
