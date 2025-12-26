package schema

import (
	"net/mail"
	"regexp"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/history"

	"github.com/theopenlane/common/models"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/token"
)

// Subscriber holds the schema definition for the Subscriber entity
type Subscriber struct {
	SchemaFuncs

	ent.Schema
}

// SchemaSubscriber is the name of the Subscriber schema.
const SchemaSubscriber = "subscriber"

// Name returns the name of the Subscriber schema.
func (Subscriber) Name() string {
	return SchemaSubscriber
}

// GetType returns the type of the Subscriber schema.
func (Subscriber) GetType() any {
	return Subscriber.Type
}

// PluralName returns the plural name of the Subscriber schema.
func (Subscriber) PluralName() string {
	return pluralize.NewClient().Plural(SchemaSubscriber)
}

// Fields of the Subscriber
func (Subscriber) Fields() []ent.Field {
	return []ent.Field{
		field.String("email").
			Comment("email address of the subscriber").
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("email"),
			).
			Validate(func(email string) error {
				_, err := mail.ParseAddress(email)
				return err
			}),
		field.String("phone_number").
			Comment("phone number of the subscriber").
			Optional().
			Validate(func(phone string) error {
				regex := `^\+[1-9]{1}[0-9]{3,14}$`
				_, err := regexp.MatchString(regex, phone)

				return err
			}),
		field.Bool("verified_email").
			Comment("indicates if the email address has been verified").
			Default(false).
			Annotations(entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput)),
		field.Bool("verified_phone").
			Comment("indicates if the phone number has been verified").
			Default(false).
			Annotations(entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput)),
		field.Bool("active").
			Comment("indicates if the subscriber is active or not, active users will have at least one verified contact method").
			Default(false).
			Annotations(entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput), entgql.OrderField("active")),
		field.String("token").
			Comment("the verification token sent to the user via email which should only be provided to the /subscribe endpoint + handler").
			Unique().
			Annotations(entgql.Skip()).
			NotEmpty(),
		field.Time("ttl").
			Comment("the ttl of the verification token which defaults to 7 days").
			Annotations(entgql.Skip()).
			Nillable(),
		field.Bytes("secret").
			Comment("the comparison secret to verify the token's signature").
			NotEmpty().
			Annotations(entgql.Skip()).
			Nillable(),
		field.Bool("unsubscribed").
			Comment("indicates if the subscriber has unsubscribed from communications").
			Default(false).
			Annotations(entgql.Skip(entgql.SkipMutationCreateInput), entgql.OrderField("unsubscribed")),
		field.Int("send_attempts").
			Comment("the number of attempts made to perform email send of the subscription, maximum of 5").
			Annotations(
				entgql.OrderField("send_attempts"),
				entgql.Skip(entgql.SkipMutationUpdateInput, entgql.SkipMutationCreateInput),
			).
			Default(1),
	}
}

// Mixin of the Subscriber
func (s Subscriber) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(s,
				withSkipTokenTypesObjects(&token.VerifyToken{}, &token.SignUpToken{}), withSkipForSystemAdmin(true)),
		},
	}.getMixins(s)
}

// Edges of the Subscriber
func (s Subscriber) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeToWithPagination(s, Event{}),
	}
}

func (Subscriber) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookSubscriberCreate(),
		hooks.HookSubscriberUpdated(),
	}
}

// Indexes of the Subscriber
func (Subscriber) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email", ownerFieldName).
			Unique().
			Annotations(
				entsql.IndexWhere("deleted_at is NULL and unsubscribed = false"),
			),
	}
}

// Annotations of the Subscriber
func (Subscriber) Annotations() []schema.Annotation {
	return []schema.Annotation{
		history.Annotations{
			Exclude: true,
		},
		entx.Exportable{},
	}
}

// Policy of the Subscriber
func (Subscriber) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			rule.AllowIfContextHasPrivacyTokenOfType[*token.SignUpToken](),
			rule.AllowIfContextHasPrivacyTokenOfType[*token.VerifyToken](),
		),
		policy.WithMutationRules(
			rule.AllowIfContextHasPrivacyTokenOfType[*token.SignUpToken](),
			rule.AllowIfContextHasPrivacyTokenOfType[*token.VerifyToken](),
			policy.CheckOrgWriteAccess(),
		),
	)
}

func (Subscriber) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogBaseModule,
	}
}
