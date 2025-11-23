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

	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/pkg/models"
)

// APIToken holds the schema definition for the APIToken entity.
type APIToken struct {
	SchemaFuncs

	ent.Schema
}

// SchemaAPIToken is the name of the APIToken schema.
const SchemaAPIToken = "api_token"

// Name returns the name of the APIToken schema.
func (APIToken) Name() string {
	return SchemaAPIToken
}

// GetType returns the type of the APIToken schema.
func (APIToken) GetType() any {
	return APIToken.Type
}

// PluralName returns the plural name of the APIToken schema.
func (APIToken) PluralName() string {
	return pluralize.NewClient().Plural(SchemaAPIToken)
}

// Fields of the APIToken
func (APIToken) Fields() []ent.Field {
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
				token := keygen.PrefixedSecret("tola") // api token prefix
				return token
			}),
		field.Time("expires_at").
			Comment("when the token expires").
			Annotations(
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
			Optional().
			Annotations(
				entgql.OrderField("last_used_at"),
			).
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
		field.JSON("sso_authorizations", models.SSOAuthorizationMap{}).
			Comment("SSO verification time for the owning organization").
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput | entgql.SkipMutationUpdateInput),
			).
			Optional(),
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
	}.getMixins(a)
}

func (APIToken) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogBaseModule,
	}
}

// Annotations of the APIToken
func (a APIToken) Annotations() []schema.Annotation {
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
func (a APIToken) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorAPIToken(),
	}
}

// Policy of the APIToken
func (a APIToken) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			policy.CheckAPITokenQueryAccess(),
		),
		policy.WithMutationRules(
			rule.RequirePaymentMethod(),
			rule.AllowIfContextAllowRule(),
			policy.CheckOrgWriteAccess(),
		),
	)
}
