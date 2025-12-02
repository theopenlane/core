package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/history"
	emixin "github.com/theopenlane/entx/mixin"

	"github.com/theopenlane/ent/hooks"
	"github.com/theopenlane/ent/interceptors"
	"github.com/theopenlane/ent/mixin"
	"github.com/theopenlane/ent/privacy/policy"
	"github.com/theopenlane/ent/privacy/rule"
	"github.com/theopenlane/ent/privacy/token"
)

// FileDownloadToken holds the schema definition for the FileDownloadToken entity
type FileDownloadToken struct {
	SchemaFuncs

	ent.Schema
}

// SchemaFileDownloadToken is the name of the FileDownloadToken schema.
const SchemaFileDownloadToken = "file_download_token"

// Name returns the name of the FileDownloadToken schema.
func (FileDownloadToken) Name() string {
	return SchemaFileDownloadToken
}

// GetType returns the type of the FileDownloadToken schema.
func (FileDownloadToken) GetType() any {
	return FileDownloadToken.Type
}

// PluralName returns the plural name of the FileDownloadToken schema.
func (FileDownloadToken) PluralName() string {
	return pluralize.NewClient().Plural(SchemaFileDownloadToken)
}

// Fields of the FileDownloadToken
func (FileDownloadToken) Fields() []ent.Field {
	return []ent.Field{
		field.String("token").
			Comment("the reset token sent to the user via email which should only be provided to the /forgot-password endpoint + handler").
			Optional().
			Unique(),
		field.Time("ttl").
			Comment("the ttl of the reset token which defaults to 15 minutes").
			Optional().
			Nillable(),
		field.String("user_id").
			Comment("the email used as input to generate the reset token; this is used to verify that the token when regenerated within the server matches the token").
			Optional().
			Nillable(),
		field.String("organization_id").
			Comment("the email used as input to generate the reset token; this is used to verify that the token when regenerated within the server matches the token").
			Optional().
			Nillable(),
		field.String("file_id").
			Comment("the email used as input to generate the reset token; this is used to verify that the token when regenerated within the server matches the token").
			Optional().
			Nillable(),
		field.Bytes("secret").
			Comment("the comparison secret to verify the token's signature").
			Sensitive().
			Optional().
			Nillable(),
	}
}

// Mixin of the FileDownloadToken
func (p FileDownloadToken) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		emixin.IDMixin{},
		mixin.SoftDeleteMixin{},
		newUserOwnedMixin(p,
			withSkipInterceptor(interceptors.SkipAll),
			withSkipTokenTypesUsers(&token.DownloadToken{}),
		),
	}
}

// Indexes of the FileDownloadToken
func (FileDownloadToken) Indexes() []ent.Index {
	return []ent.Index{
		// FileDownloadToken should be unique, but ignore deleted FileDownloadTokens
		index.Fields("token").
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Annotations of the FileDownloadToken
func (p FileDownloadToken) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.Skip(entgql.SkipAll),
		entx.SchemaGenSkip(true),
		entx.QueryGenSkip(true),
		history.Annotations{
			Exclude: true,
		},
	}
}

// Hooks of the FileDownloadToken
func (FileDownloadToken) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookFileDownloadToken(),
	}
}

// Policy of the FileDownloadToken
func (p FileDownloadToken) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			rule.AllowAfterApplyingPrivacyTokenFilter[*token.DownloadToken](),
		),
		policy.WithOnMutationRules(
			ent.OpCreate,
			rule.AllowIfContextHasPrivacyTokenOfType[*token.DownloadToken](),
			rule.AllowMutationAfterApplyingOwnerFilter(),
		),
		policy.WithOnMutationRules(
			ent.OpUpdateOne|ent.OpUpdate|ent.OpDeleteOne|ent.OpDelete,
			rule.AllowMutationAfterApplyingOwnerFilter(),
		),
	)
}
