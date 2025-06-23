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
	"github.com/theopenlane/entx"

	"github.com/theopenlane/entx/history"
	emixin "github.com/theopenlane/entx/mixin"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/pkg/enums"
)

// TemplateRecipient holds the schema definition for the EmailVerificationToken entity
type TemplateRecipient struct {
	SchemaFuncs

	ent.Schema
}

// SchemaTemplateRecipient is the name of the EmailVerificationToken schema.
const SchemaTemplateRecipient = "template_recipient" // nolint:gosec

// Name returns the name of the TemplateRecipient schema.
func (TemplateRecipient) Name() string {
	return SchemaTemplateRecipient
}

// GetType returns the type of the TemplateRecipient schema.
func (TemplateRecipient) GetType() any {
	return TemplateRecipient.Type
}

// PluralName returns the plural name of the TemplateRecipient schema.
func (TemplateRecipient) PluralName() string {
	return pluralize.NewClient().Plural(SchemaTemplateRecipient)
}

// Fields of the TemplateRecipient
func (TemplateRecipient) Fields() []ent.Field {
	return []ent.Field{

		field.String("token").
			Comment("the verification token sent to the user via email").
			Unique().
			NotEmpty(),

		field.Time("ttl").
			Comment("the ttl for the survey to be filled").
			Nillable(),

		field.String("email").
			Comment("the recipient email for the questionairre").
			Validate(func(email string) error {
				_, err := mail.ParseAddress(email)
				return err
			}).
			NotEmpty(),

		field.Bytes("secret").
			Comment("the comparison secret to verify the token's signature"),

		field.String("template_id").
			Comment("the ID of the template this token belongs to"),

		field.Int("send_attempts").
			Comment("the number of attempts made to send the questionairre to the user, maximum of 5").
			Annotations(
				entgql.OrderField("send_attempts"),
				entgql.Skip(entgql.SkipMutationUpdateInput, entgql.SkipMutationCreateInput),
			).
			Default(1),

		field.Enum("status").
			GoType(enums.TemplateRecipientStatus("")).
			Default(enums.TemplateRecipientStatusActive.String()).
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput | entgql.SkipMutationUpdateInput),
			).
			Comment("the status of this token. Defaults to active"),

		field.String("document_data_id").
			Optional().
			Comment("the ID of the document this recipient belongs to. This will only be available if the survey was ever filled"),
	}
}

// Edges of the TemplateRecipient
func (e TemplateRecipient) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: e,
			edgeSchema: DocumentData{},
			field:      "document_data_id",
			name:       "document_data",
			required:   false,
		}),

		uniqueEdgeTo(&edgeDefinition{
			fromSchema: e,
			edgeSchema: Template{},
			field:      "template_id",
			name:       "template_id",
			required:   true,
		}),
	}
}

// Mixin of the TemplateRecipient
func (e TemplateRecipient) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		emixin.IDMixin{},
		mixin.SoftDeleteMixin{},
	}
}

// Indexes of the TemplateRecipient
func (TemplateRecipient) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("token").
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Annotations of the TemplateRecipient
func (TemplateRecipient) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entx.Features("base"),
		entgql.Skip(entgql.SkipAll),
		entx.SchemaGenSkip(true),
		entx.QueryGenSkip(true),
		history.Annotations{
			Exclude: true,
		},
	}
}

// Hooks of the TemplateRecipient
func (TemplateRecipient) Hooks() []ent.Hook {
	return []ent.Hook{}
}

// Policy of the TemplateRecipient
func (TemplateRecipient) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(),
		),
		policy.WithOnMutationRules(
			ent.OpCreate,
			privacy.AlwaysAllowRule(),
		),
	)
}
