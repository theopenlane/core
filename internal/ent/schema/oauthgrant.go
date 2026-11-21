package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx/history"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/ent/privacy/policy"
	"github.com/theopenlane/core/v2/internal/ent/privacy/rule"
)

// OauthGrant holds the schema definition for a third-party application installed through the OAuth consent screen
type OauthGrant struct {
	SchemaFuncs

	ent.Schema
}

// SchemaOauthGrant is the name of the OauthGrant schema.
const SchemaOauthGrant = "oauth_grant"

// Name returns the name of the OauthGrant schema.
func (OauthGrant) Name() string {
	return SchemaOauthGrant
}

// GetType returns the type of the OauthGrant schema.
func (OauthGrant) GetType() any {
	return OauthGrant.Type
}

// PluralName returns the plural name of the OauthGrant schema.
func (OauthGrant) PluralName() string {
	return pluralize.NewClient().Plural(SchemaOauthGrant)
}

// Fields of the OauthGrant
func (OauthGrant) Fields() []ent.Field {
	return []ent.Field{
		field.String("client_id").
			Comment("the OAuth client_id of the installed application, the URL of its client metadata document").
			Immutable().
			NotEmpty().
			Annotations(entgql.Skip(entgql.SkipMutationCreateInput | entgql.SkipMutationUpdateInput)),
		field.String("client_name").
			Comment("the application's display name from its client metadata document").
			NotEmpty().
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput|entgql.SkipMutationUpdateInput),
				entgql.OrderField("client_name"),
			),
		field.String("client_uri").
			Comment("the application's home page").
			Optional().
			Annotations(entgql.Skip(entgql.SkipMutationCreateInput | entgql.SkipMutationUpdateInput)),
		field.String("logo_uri").
			Comment("the application's logo").
			Optional().
			Annotations(entgql.Skip(entgql.SkipMutationCreateInput | entgql.SkipMutationUpdateInput)),
		field.Enum("kind").
			Comment("USER for a personal connection that acts as the user, ORGANIZATION for an install that acts as the organization").
			GoType(enums.OAuthGrantKind("")).
			Default(enums.OAuthGrantKindUser.String()).
			Immutable().
			Annotations(entgql.Skip(entgql.SkipMutationCreateInput | entgql.SkipMutationUpdateInput)),
		field.String("user_id").
			Comment("for USER grants, the user the application acts as").
			Optional().
			Immutable().
			Annotations(entgql.Skip(entgql.SkipMutationCreateInput | entgql.SkipMutationUpdateInput)),
		field.Enum("access_level").
			Comment("for USER grants, READ or WRITE, the application never exceeds the user's own permissions").
			GoType(enums.OAuthAccessLevel("")).
			Default(enums.OAuthAccessLevelRead.String()).
			Optional().
			Annotations(entgql.Skip(entgql.SkipMutationCreateInput | entgql.SkipMutationUpdateInput)),
		field.JSON("scopes", []string{}).
			Comment("the scopes granted, object:action scopes for ORGANIZATION grants and personal scopes for USER grants").
			Optional().
			Annotations(entgql.Skip(entgql.SkipMutationCreateInput | entgql.SkipMutationUpdateInput)),
		field.String("installed_by_user_id").
			Comment("the user who approved the application").
			Immutable().
			Annotations(entgql.Skip(entgql.SkipMutationCreateInput | entgql.SkipMutationUpdateInput)),
		field.Time("last_used_at").
			Comment("when an access token under this grant was last used").
			Optional().
			Nillable().
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput|entgql.SkipMutationUpdateInput),
				entgql.OrderField("last_used_at"),
			),
		field.Bool("is_active").
			Comment("whether the grant is active, set to false to revoke the application's access").
			Default(true).
			Annotations(
				entgql.OrderField("is_active"),
			),
		field.String("revoked_reason").
			Comment("the reason the grant was revoked").
			Optional().
			Nillable(),
		field.String("revoked_by").
			Comment("the user who revoked the grant").
			Optional().
			Nillable().
			Annotations(entgql.Skip(entgql.SkipMutationCreateInput | entgql.SkipMutationUpdateInput)),
		field.Time("revoked_at").
			Comment("when the grant was revoked").
			Optional().
			Nillable().
			Annotations(entgql.Skip(entgql.SkipMutationCreateInput | entgql.SkipMutationUpdateInput)),
	}
}

// Edges of the OauthGrant
func (g OauthGrant) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: g,
			edgeSchema: User{},
			field:      "user_id",
			immutable:  true,
			comment:    "for USER grants, the user the application acts as",
			annotations: []schema.Annotation{
				entgql.Skip(entgql.SkipMutationCreateInput | entgql.SkipMutationUpdateInput),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:    g,
			edgeSchema:    OauthToken{},
			cascadeDelete: "OauthGrant",
			comment:       "authorization codes, access tokens, and refresh tokens issued under this grant",
			annotations: []schema.Annotation{
				entgql.Skip(entgql.SkipAll),
			},
		}),
	}
}

// Indexes of the OauthGrant
func (OauthGrant) Indexes() []ent.Index {
	return []ent.Index{
		// one organization install per application, and one personal connection per application per user
		index.Fields("client_id", ownerFieldName, "kind", "user_id").
			Unique().
			Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Mixin of the OauthGrant
func (g OauthGrant) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(g),
		},
	}.getMixins(g)
}

// Modules of the OauthGrant
func (OauthGrant) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogAnyModule,
	}
}

// Annotations of the OauthGrant
func (OauthGrant) Annotations() []schema.Annotation {
	return []schema.Annotation{
		history.Annotations{
			Exclude: true,
		},
	}
}

// Hooks of the OauthGrant
func (OauthGrant) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookCreateOauthGrant(),
		hooks.HookUpdateOauthGrant(),
		hooks.HookDeleteOauthGrant(),
	}
}

// Policy of the OauthGrant
func (OauthGrant) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithDenyAppMutations(),
		// only the consent flow creates grants, so a scoped token can't mint one
		policy.WithOnMutationRulesBeforeScope(ent.OpCreate,
			privacy.AlwaysDenyRule(),
		),
		policy.WithMutationRules(
			rule.AllowIfOwnOauthGrant(),
			policy.CheckOrgWriteAccess(),
		),
	)
}
