package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx/history"

	"github.com/theopenlane/utils/keygen"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
)

// PersonalAccessToken holds the schema definition for the PersonalAccessToken entity.
type PersonalAccessToken struct {
	SchemaFuncs

	ent.Schema
}

const SchemaPersonalAccessToken = "personal_access_token"

func (PersonalAccessToken) Name() string {
	return SchemaPersonalAccessToken
}

func (PersonalAccessToken) GetType() any {
	return PersonalAccessToken.Type
}

func (PersonalAccessToken) PluralName() string {
	return pluralize.NewClient().Plural(SchemaPersonalAccessToken)
}

// Fields of the PersonalAccessToken
func (PersonalAccessToken) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name associated with the token").
			Annotations(
				entgql.OrderField("name"),
			).
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
				entgql.OrderField("expires_at"),
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
			Annotations(
				entgql.OrderField("last_used_at"),
			).
			Optional().
			Nillable(),
		field.Bool("is_active").
			Default(true).
			Comment("whether the token is active").
			Annotations(
				entgql.OrderField("is_active"),
			).
			Optional(),
		field.String("revoked_reason").
			Comment("the reason the token was revoked").
			Annotations(
				entgql.Skip(entgql.SkipMutationUpdateInput, entgql.SkipMutationCreateInput),
			).
			Optional().
			Nillable(),
		field.String("revoked_by").
			Comment("the user who revoked the token").
			Annotations(
				entgql.Skip(entgql.SkipMutationUpdateInput, entgql.SkipMutationCreateInput),
			).
			Optional().
			Nillable(),
		field.Time("revoked_at").
			Comment("when the token was revoked").
			Annotations(
				entgql.Skip(entgql.SkipMutationUpdateInput, entgql.SkipMutationCreateInput),
			).
			Optional().
			Nillable(),
	}
}

// Edges of the PersonalAccessToken
func (p PersonalAccessToken) Edges() []ent.Edge {
	return []ent.Edge{
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: p,
			edgeSchema: Organization{},
			comment:    "the organization(s) the token is associated with",
		}),
		defaultEdgeToWithPagination(p, Event{}),
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
func (p PersonalAccessToken) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newUserOwnedMixin(p,
				withSkipInterceptor(interceptors.SkipOnlyQuery),
			),
		},
	}.getMixins()
}

// Annotations of the PersonalAccessToken
func (PersonalAccessToken) Annotations() []schema.Annotation {
	return []schema.Annotation{
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
