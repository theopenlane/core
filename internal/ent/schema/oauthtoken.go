package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/history"
	iamoauth "github.com/theopenlane/iam/oauthserver"

	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/privacy/policy"
)

// OauthToken holds the schema definition for OAuth authorization requests, codes, access tokens, and refresh tokens
// only a keyed HMAC of each token is stored, and the row ID is the token's ID
type OauthToken struct {
	SchemaFuncs

	ent.Schema
}

// SchemaOauthToken is the name of the OauthToken schema.
const SchemaOauthToken = "oauth_token" //nolint:gosec

// Name returns the name of the OauthToken schema.
func (OauthToken) Name() string {
	return SchemaOauthToken
}

// GetType returns the type of the OauthToken schema.
func (OauthToken) GetType() any {
	return OauthToken.Type
}

// PluralName returns the plural name of the OauthToken schema.
func (OauthToken) PluralName() string {
	return pluralize.NewClient().Plural(SchemaOauthToken)
}

// Fields of the OauthToken
func (OauthToken) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("kind").
			Comment("what the token is, a pending authorization request, a code, an access token, or a refresh token").
			GoType(iamoauth.TokenKind("")).
			Default(string(iamoauth.KindAccess)).
			Immutable(),
		field.String("token_hash").
			Comment("keyed HMAC of the token from the API token keyring, the raw token is never stored").
			Unique().
			Immutable().
			Sensitive(),
		field.String("key_version").
			Comment("the API token keyring version the token was hashed with, so keys can rotate").
			Immutable().
			NotEmpty(),
		field.Time("expires_at").
			Comment("when the token expires").
			Immutable(),
		field.Time("consumed_at").
			Comment("when a single-use request, code, or refresh token was used, or when the token was revoked").
			Optional().
			Nillable(),
		field.String("client_id").
			Comment("the client_id of the application the token was issued to").
			Immutable().
			NotEmpty(),
		field.String("redirect_uri").
			Comment("the redirect_uri of the authorization request, which the token request must repeat").
			Optional().
			Immutable(),
		field.String("code_challenge").
			Comment("the PKCE S256 code_challenge of the authorization request").
			Optional().
			Immutable(),
		field.String("state").
			Comment("the client's state parameter, echoed back on the redirect").
			Optional().
			Immutable(),
		field.String("resource").
			Comment("the RFC 8707 resource indicator the token is bound to, such as the MCP server URL").
			Optional().
			Immutable(),
		field.JSON("scopes", []string{}).
			Comment("the scopes requested for a REQUEST, or granted for a CODE, ACCESS, or REFRESH").
			Optional(),
		field.String("user_id").
			Comment("the user who consented").
			Optional(),
		field.String("oauth_grant_id").
			Comment("the grant the token was issued under, empty for pending requests").
			Optional(),
	}
}

// Edges of the OauthToken
func (t OauthToken) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: t,
			edgeSchema: OauthGrant{},
			field:      "oauth_grant_id",
		}),
	}
}

// Indexes of the OauthToken
func (OauthToken) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("oauth_grant_id", "kind"),
		index.Fields("expires_at"),
	}
}

// Mixin of the OauthToken
func (t OauthToken) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeSoftDelete:  true,
		excludeTags:        true,
		excludeAnnotations: true,
	}.getMixins(t)
}

// Annotations of the OauthToken
func (OauthToken) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.Skip(entgql.SkipAll),
		entx.SchemaGenSkip(true),
		entx.QueryGenSkip(true),
		history.Annotations{
			Exclude: true,
		},
	}
}

// Policy of the OauthToken
func (OauthToken) Policy() ent.Policy {
	// internal operations are allowed by the pre-policy, everything else is denied
	return policy.NewPolicy(
		policy.WithSkipDenyOrganizationRule(),
		policy.WithQueryRules(
			privacy.AlwaysDenyRule(),
		),
		policy.WithMutationRulesBeforeScope(
			privacy.AlwaysDenyRule(),
		),
	)
}
