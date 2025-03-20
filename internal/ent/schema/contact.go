package schema

import (
	"net/mail"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/entfga"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/validator"
	"github.com/theopenlane/core/pkg/enums"
)

// Contact holds the schema definition for the Contact entity
type Contact struct {
	SchemaFuncs

	ent.Schema
}

const SchemaContact = "contact"

func (Contact) Name() string {
	return SchemaContact
}

func (Contact) GetType() any {
	return Contact.Type
}

func (Contact) PluralName() string {
	return pluralize.NewClient().Plural(SchemaContact)
}

// Fields of the Contact
func (Contact) Fields() []ent.Field {
	return []ent.Field{
		field.String("full_name").
			Comment("the full name of the contact").
			MaxLen(nameMaxLen).
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("full_name"),
			).
			NotEmpty(),
		field.String("title").
			Comment("the title of the contact").
			Annotations(
				entgql.OrderField("title"),
			).
			Optional(),
		field.String("company").
			Comment("the company of the contact").
			Annotations(
				entgql.OrderField("company"),
			).
			Optional(),
		field.String("email").
			Comment("the email of the contact").
			Optional().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("email"),
			).
			Validate(func(email string) error {
				_, err := mail.ParseAddress(email)
				return err
			}),
		field.String("phone_number").
			Comment("the phone number of the contact").
			Validate(func(s string) error {
				valid := validator.ValidatePhoneNumber(s)
				if !valid {
					return rout.InvalidField("phone_number")
				}

				return nil
			}).
			Optional(),
		field.String("address").
			Comment("the address of the contact").
			Optional(),
		field.Enum("status").
			Comment("status of the contact").
			Annotations(
				entgql.OrderField("STATUS"),
			).
			GoType(enums.UserStatus("")).
			Default(enums.UserStatusActive.String()),
	}
}

// Mixin of the Contact
func (c Contact) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			NewOrgOwnMixinWithRef(c.PluralName()),
		},
	}.getMixins()
}

// Edges of the Contact
func (c Contact) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeFromWithPagination(c, Entity{}),
		defaultEdgeToWithPagination(c, File{}),
	}
}

// Annotations of the Contact
func (Contact) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.OrganizationInheritedChecks(),
	}
}

// Hooks of the Contact
func (Contact) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookContact(),
	}
}

// Policy of the Contact
func (Contact) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(), //  interceptor should filter out the results
		),
		policy.WithMutationRules(
			entfga.CheckEditAccess[*generated.ContactMutation](),
		),
	)
}
