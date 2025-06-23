package schema

import (
	"net/mail"
	"time"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/history"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/pkg/enums"
)

const (
	defaultTemplateExpiration = time.Hour * 168
)

// TemplateRecipient holds the schema definition for the EmailVerificationToken entity
type TemplateRecipient struct {
	SchemaFuncs

	ent.Schema
}

// SchemaTemplateRecipient is the name of the EmailVerificationToken schema.
const SchemaTemplateRecipient = "template_recipient"

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
			NotEmpty().
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput),
			).
			Immutable(),

		field.Time("expires_at").
			Comment("when the token expires").
			Annotations(
				entgql.OrderField("expires_at"),
				entgql.Skip(entgql.SkipMutationUpdateInput|entgql.SkipMutationCreateInput),
			).
			Default(time.Now().Add(defaultTemplateExpiration)).
			Immutable(),

		field.String("email").
			Comment("the recipient email for the questionairre").
			Validate(func(email string) error {
				_, err := mail.ParseAddress(email)
				return err
			}).
			NotEmpty().
			Immutable(),

		field.String("secret").
			Comment("the comparison secret to verify the token's signature").
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput),
			).
			Immutable(),

		field.String("template_id").
			Comment("the ID of the template this token belongs to").
			Immutable(),

		field.Int("send_attempts").
			Comment("the number of attempts made to send the questionairre to the user, maximum of 5").
			Annotations(
				entgql.OrderField("send_attempts"),
				entgql.Skip(entgql.SkipMutationUpdateInput|entgql.SkipMutationCreateInput),
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
			Comment("the ID of the document this recipient belongs to. This will only be available if the survey was ever filled").
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput),
			),
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
			immutable:  true,
		}),

		defaultEdgeToWithPagination(e, Event{}),
	}
}

// Mixin of the TemplateRecipient
func (e TemplateRecipient) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags: true,
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(e),
		},
	}.getMixins()
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
