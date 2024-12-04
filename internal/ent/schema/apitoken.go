package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/theopenlane/entx/history"
	emixin "github.com/theopenlane/entx/mixin"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/utils/keygen"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
)

// APIToken holds the schema definition for the APIToken entity.
type APIToken struct {
	ent.Schema
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
	}
}

// Edges of the APIToken
func (APIToken) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the APIToken
func (APIToken) Indexes() []ent.Index {
	return []ent.Index{
		// non-unique index.
		index.Fields("token"),
	}
}

// Mixin of the APIToken
func (APIToken) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		mixin.SoftDeleteMixin{},
		emixin.IDMixin{},
		emixin.TagMixin{},
		NewOrgOwnedMixin(
			ObjectOwnedMixin{
				Ref: "api_tokens",
				// skip the interceptor for Only queries when the token is being checked
				// and we do not yet know the organization
				SkipInterceptor: interceptors.SkipOnlyQuery,
			}),
	}
}

// Annotations of the APIToken
func (APIToken) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entgql.RelayConnection(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
		history.Annotations{
			Exclude: true,
		},
		entfga.OrganizationInheritedChecks(),
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
			entfga.CheckReadAccess[*generated.APITokenQuery](),
		),
		policy.WithMutationRules(
			entfga.CheckEditAccess[*generated.APITokenMutation](),
		),
	)
}
