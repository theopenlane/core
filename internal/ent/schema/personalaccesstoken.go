package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/theopenlane/entx/history"
	emixin "github.com/theopenlane/entx/mixin"

	"github.com/theopenlane/utils/keygen"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
)

// PersonalAccessToken holds the schema definition for the PersonalAccessToken entity.
type PersonalAccessToken struct {
	ent.Schema
}

// Fields of the PersonalAccessToken
func (PersonalAccessToken) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name associated with the token").
			NotEmpty(),
		field.String("token").
			Unique().
			Immutable().
			Annotations(
				entgql.Skip(^entgql.SkipType),
			).
			DefaultFunc(func() string {
				token := keygen.PrefixedSecret("tolp") // token prefix
				return token
			}),
		field.Time("expires_at").
			Comment("when the token expires").
			Annotations(
				entgql.Skip(entgql.SkipMutationUpdateInput),
			).
			Optional().
			Nillable(),
		field.String("description").
			Comment("a description of the token's purpose").
			Optional().
			Nillable().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.JSON("scopes", []string{}).
			Optional(),
		field.Time("last_used_at").
			Optional().
			Nillable(),
		field.Bool("is_active").
			Default(true).
			Comment("whether the token is active").
			Optional(),
		field.String("revoked_reason").
			Comment("the reason the token was revoked").
			Optional().
			Nillable(),
		field.String("revoked_by").
			Comment("the user who revoked the token").
			Optional().
			Nillable(),
		field.Time("revoked_at").
			Comment("when the token was revoked").
			Optional().
			Nillable(),
	}
}

// Edges of the PersonalAccessToken
func (PersonalAccessToken) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("organizations", Organization.Type).
			Ref("personal_access_tokens").
			Annotations(entgql.RelayConnection()).
			Comment("the organization(s) the token is associated with"),
		edge.To("events", Event.Type).Annotations(entgql.RelayConnection()),
	}
}

// Indexes of the PersonalAccessToken
func (PersonalAccessToken) Indexes() []ent.Index {
	return []ent.Index{
		// non-unique index.
		index.Fields("token"),
	}
}

// Mixin of the PersonalAccessToken
func (PersonalAccessToken) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		mixin.SoftDeleteMixin{},
		emixin.IDMixin{},
		emixin.TagMixin{},
		UserOwnedMixin{
			Ref:         "personal_access_tokens",
			AllowUpdate: false,
			// skip the interceptor for Only queries when the token is being checked
			// and we do not yet know the organization
			SkipInterceptor: interceptors.SkipOnlyQuery,
		},
	}
}

// Annotations of the PersonalAccessToken
func (PersonalAccessToken) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entgql.RelayConnection(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
		history.Annotations{
			Exclude: true,
		},
	}
}

// Hooks of the PersonalAccessToken
func (PersonalAccessToken) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookCreatePersonalAccessToken(),
		hooks.HookUpdatePersonalAccessToken(),
	}
}

// Interceptors of the PersonalAccessToken
func (PersonalAccessToken) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorPat(),
	}
}

// Policy of the PersonalAccessToken
func (PersonalAccessToken) Policy() ent.Policy {
	return privacy.Policy{
		Mutation: privacy.MutationPolicy{
			rule.AllowIfContextAllowRule(),
			rule.AllowMutationAfterApplyingOwnerFilter(),
			privacy.AlwaysAllowRule(),
		},
		Query: privacy.QueryPolicy{
			privacy.AlwaysAllowRule(),
		},
	}
}
