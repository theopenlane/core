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
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
)

// APIToken holds the schema definition for the APIToken entity.
type APIToken struct {
	SchemaFuncs

	ent.Schema
}

const SchemaAPIToken = "api_token"

func (APIToken) Name() string {
	return SchemaAPIToken
}

func (APIToken) GetType() any {
	return APIToken.Type
}

func (APIToken) PluralName() string {
	return pluralize.NewClient().Plural(SchemaAPIToken)
}

// Fields of the APIToken
func (APIToken) Fields() []ent.Field {
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
				token := keygen.PrefixedSecret("tola") // api token prefix
				return token
			}),
		field.Time("expires_at").
			Comment("when the token expires").
			Annotations(
				entgql.OrderField("expires_at"),
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
			Annotations(
				entgql.OrderField("last_used_at"),
			).
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

// Indexes of the APIToken
func (APIToken) Indexes() []ent.Index {
	return []ent.Index{
		// non-unique index.
		index.Fields("token"),
	}
}

// Mixin of the APIToken
func (a APIToken) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(a),
		},
	}.getMixins()
}

// Annotations of the APIToken
func (APIToken) Annotations() []schema.Annotation {
	return []schema.Annotation{
		history.Annotations{
			Exclude: true,
		},
	}
}

// Hooks of the APIToken
func (APIToken) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookCreateAPIToken(),
		hooks.HookUpdateAPIToken(),
	}
}

// Interceptors of the APIToken
func (APIToken) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorAPIToken(),
	}
}

// Policy of the APIToken
func (APIToken) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(), //  interceptor should filter out the results
		),
		policy.WithMutationRules(
			rule.AllowIfContextAllowRule(),
			policy.CheckOrgWriteAccess(),
		),
	)
}
