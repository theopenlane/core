package schema

import (
	"net/mail"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"

	"github.com/theopenlane/entx"
	emixin "github.com/theopenlane/entx/mixin"
	"github.com/theopenlane/iam/entfga"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/validator"
	"github.com/theopenlane/core/pkg/enums"
)

// Contact holds the schema definition for the Contact entity
type Contact struct {
	ent.Schema
}

// Fields of the Contact
func (Contact) Fields() []ent.Field {
	return []ent.Field{
		field.String("full_name").
			Comment("the full name of the contact").
			MaxLen(nameMaxLen).
			Annotations(
				entx.FieldSearchable(),
			).
			NotEmpty(),
		field.String("title").
			Comment("the title of the contact").
			Optional(),
		field.String("company").
			Comment("the company of the contact").
			Optional(),
		field.String("email").
			Comment("the email of the contact").
			Optional().
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
			GoType(enums.UserStatus("")).
			Default(enums.UserStatusActive.String()),
	}
}

// Mixin of the Contact
func (Contact) Mixin() []ent.Mixin {
	return []ent.Mixin{
		NewAuditMixin(),
		emixin.IDMixin{},
		mixin.SoftDeleteMixin{},
		emixin.TagMixin{},
		NewOrgOwnMixinWithRef("contacts"),
	}
}

// Edges of the Contact
func (Contact) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("entities", Entity.Type).
			Ref("contacts"),
		edge.To("files", File.Type),
	}
}

// Indexes of the Contact
func (Contact) Indexes() []ent.Index {
	return []ent.Index{}
}

// Annotations of the Contact
func (Contact) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entgql.RelayConnection(),
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entfga.OrganizationInheritedChecks(),
	}
}

// Hooks of the Contact
func (Contact) Hooks() []ent.Hook {
	return []ent.Hook{}
}

// Interceptors of the Contact
func (Contact) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}

// Policy of the Contact
func (Contact) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			entfga.CheckReadAccess[*generated.ContactQuery](),
		),
		policy.WithMutationRules(
			entfga.CheckEditAccess[*generated.ContactMutation](),
		),
	)
}
