package schema

import (
	"net/mail"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/common/models"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/validator"
)

// Contact holds the schema definition for the Contact entity
type Contact struct {
	SchemaFuncs

	ent.Schema
}

// SchemaContact is the name of the Contact schema.
const SchemaContact = "contact"

// Name returns the name of the Contact schema.
func (Contact) Name() string {
	return SchemaContact
}

// GetType returns the type of the Contact schema.
func (Contact) GetType() any {
	return Contact.Type
}

// PluralName returns the plural name of the Contact schema.
func (Contact) PluralName() string {
	return pluralize.NewClient().Plural(SchemaContact)
}

// Fields of the Contact
func (Contact) Fields() []ent.Field {
	return []ent.Field{
		field.String("full_name").
			Comment("the full name of the contact").
			Optional().
			MaxLen(nameMaxLen).
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("full_name"),
			),
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
			newOrgOwnedMixin(c),
		},
	}.getMixins(c)
}

// Edges of the Contact
func (c Contact) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeFromWithPagination(c, Entity{}),
		defaultEdgeToWithPagination(c, File{}),
	}
}

// Hooks of the Contact
func (Contact) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookContact(),
	}
}

// Policy of the Contact
func (c Contact) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckOrgWriteAccess(),
		),
	)
}

func (Contact) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogEntityManagementModule,
	}
}

// Annotations of the Contact
func (c Contact) Annotations() []schema.Annotation {
	return []schema.Annotation{}
}
