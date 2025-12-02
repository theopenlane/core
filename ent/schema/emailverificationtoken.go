package schema

import (
	"net/mail"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/core/pkg/models"
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

// EmailVerificationToken holds the schema definition for the EmailVerificationToken entity
type EmailVerificationToken struct {
	SchemaFuncs

	ent.Schema
}

// SchemaEmailVerificationToken is the name of the EmailVerificationToken schema.
const SchemaEmailVerificationToken = "email_verification_token" // nolint:gosec

// Name returns the name of the EmailVerificationToken schema.
func (EmailVerificationToken) Name() string {
	return SchemaEmailVerificationToken
}

// GetType returns the type of the EmailVerificationToken schema.
func (EmailVerificationToken) GetType() any {
	return EmailVerificationToken.Type
}

// PluralName returns the plural name of the EmailVerificationToken schema.
func (EmailVerificationToken) PluralName() string {
	return pluralize.NewClient().Plural(SchemaEmailVerificationToken)
}

// Fields of the EmailVerificationToken
func (EmailVerificationToken) Fields() []ent.Field {
	return []ent.Field{
		field.String("token").
			Comment("the verification token sent to the user via email which should only be provided to the /verify endpoint + handler").
			Unique().
			NotEmpty(),
		field.Time("ttl").
			Comment("the ttl of the verification token which defaults to 7 days").
			Nillable(),
		field.String("email").
			Comment("the email used as input to generate the verification token; this is used to verify that the token when regenerated within the server matches the token emailed").
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

// Edges of the EmailVerificationToken
func (EmailVerificationToken) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Mixin of the EmailVerificationToken
func (e EmailVerificationToken) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		emixin.IDMixin{},
		mixin.SoftDeleteMixin{},
		newUserOwnedMixin(e,
			withSkipInterceptor(interceptors.SkipAll),
			withSkipTokenTypesUsers(&token.VerifyToken{}),
		),
	}
}

func (EmailVerificationToken) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogBaseModule,
	}
}

// Indexes of the EmailVerificationToken
func (EmailVerificationToken) Indexes() []ent.Index {
	return []ent.Index{
		// EmailVerificationTokens should be unique, but ignore deleted EmailVerificationTokens
		index.Fields("token").
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Annotations of the EmailVerificationToken
func (e EmailVerificationToken) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.Skip(entgql.SkipAll),
		entx.SchemaGenSkip(true),
		entx.QueryGenSkip(true),
		history.Annotations{
			Exclude: true,
		},
	}
}

// Hooks of the EmailVerificationToken
func (EmailVerificationToken) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookEmailVerificationToken(),
	}
}

// Policy of the EmailVerificationToken
func (e EmailVerificationToken) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			rule.AllowAfterApplyingPrivacyTokenFilter[*token.VerifyToken](),
		),
		policy.WithOnMutationRules(
			ent.OpCreate,
			rule.AllowIfContextHasPrivacyTokenOfType[*token.ResetToken](),
			rule.AllowMutationAfterApplyingOwnerFilter(),
		),
		policy.WithOnMutationRules(
			ent.OpUpdateOne|ent.OpUpdate|ent.OpDeleteOne|ent.OpDelete,
			rule.AllowMutationAfterApplyingOwnerFilter(),
		),
	)
}
