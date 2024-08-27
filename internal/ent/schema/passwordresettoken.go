package schema

import (
	"net/mail"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/entql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/datumforge/enthistory"
	"github.com/datumforge/entx"
	emixin "github.com/datumforge/entx/mixin"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/token"
)

// PasswordResetToken holds the schema definition for the PasswordResetToken entity
type PasswordResetToken struct {
	ent.Schema
}

// Fields of the PasswordResetToken
func (PasswordResetToken) Fields() []ent.Field {
	return []ent.Field{
		field.String("token").
			Comment("the reset token sent to the user via email which should only be provided to the /forgot-password endpoint + handler").
			Unique().
			NotEmpty(),
		field.Time("ttl").
			Comment("the ttl of the reset token which defaults to 15 minutes").
			Nillable(),
		field.String("email").
			Comment("the email used as input to generate the reset token; this is used to verify that the token when regenerated within the server matches the token emailed").
			Validate(func(email string) error {
				_, err := mail.ParseAddress(email)
				return err
			}).
			NotEmpty(),
		field.Bytes("secret").
			Comment("the comparison secret to verify the token's signature").
			NotEmpty().
			Nillable(),
	}
}

// Edges of the PasswordResetToken
func (PasswordResetToken) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Mixin of the PasswordResetToken
func (PasswordResetToken) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		emixin.IDMixin{},
		mixin.SoftDeleteMixin{},
		UserOwnedMixin{
			Ref:               "password_reset_tokens",
			SkipOASGeneration: true,
			SkipInterceptor:   interceptors.SkipAll,
		},
	}
}

// Indexes of the PasswordResetToken
func (PasswordResetToken) Indexes() []ent.Index {
	return []ent.Index{
		// PasswordResetTokens should be unique, but ignore deleted PasswordResetTokens
		index.Fields("token").
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Annotations of the PasswordResetToken
func (PasswordResetToken) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.Skip(entgql.SkipAll),
		entx.SchemaGenSkip(true),
		entx.QueryGenSkip(true),
		enthistory.Annotations{
			Exclude: true,
		},
	}
}

// Hooks of the PasswordResetToken
func (PasswordResetToken) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookPasswordResetToken(),
	}
}

// Policy of the PasswordResetToken
func (PasswordResetToken) Policy() ent.Policy {
	return privacy.Policy{
		Query: privacy.QueryPolicy{
			rule.AllowAfterApplyingPrivacyTokenFilter(
				&token.ResetToken{},
				func(t token.PrivacyToken, filter privacy.Filter) {
					actualToken := t.(*token.ResetToken)
					tokenFilter := filter.(*generated.PasswordResetTokenFilter)
					tokenFilter.WhereToken(entql.StringEQ(actualToken.GetToken()))
				},
			),
			privacy.AlwaysAllowRule(),
		},
		Mutation: privacy.MutationPolicy{
			privacy.OnMutationOperation(
				privacy.MutationPolicy{
					rule.AllowIfContextHasPrivacyTokenOfType(&token.ResetToken{}),
					rule.AllowMutationAfterApplyingOwnerFilter(),
					privacy.AlwaysDenyRule(),
				},
				ent.OpCreate,
			),
			privacy.OnMutationOperation(
				privacy.MutationPolicy{
					rule.AllowMutationAfterApplyingOwnerFilter(),
					privacy.AlwaysDenyRule(),
				},
				ent.OpUpdateOne|ent.OpUpdate|ent.OpDeleteOne|ent.OpDelete,
			),
		},
	}
}
