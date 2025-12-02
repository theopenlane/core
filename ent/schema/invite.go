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

	"github.com/theopenlane/shared/enums"
	"github.com/theopenlane/shared/models"
	"github.com/theopenlane/ent/hooks"
	"github.com/theopenlane/ent/privacy/policy"
	"github.com/theopenlane/ent/privacy/rule"
	"github.com/theopenlane/ent/privacy/token"
)

const (
	defaultInviteExpiresDays = 14
)

// Invite holds the schema definition for the Invite entity
type Invite struct {
	SchemaFuncs

	ent.Schema
}

const SchemaInvite = "invite"

func (Invite) Name() string {
	return SchemaInvite
}

func (Invite) GetType() any {
	return Invite.Type
}

func (Invite) PluralName() string {
	return pluralize.NewClient().Plural(SchemaInvite)
}

// Fields of the Invite
func (Invite) Fields() []ent.Field {
	return []ent.Field{
		field.String("token").
			Comment("the invitation token sent to the user via email which should only be provided to the /verify endpoint + handler").
			Unique().
			Sensitive().
			Annotations(
				entgql.Skip(),
			).
			NotEmpty(),
		field.Time("expires").
			Comment("the expiration date of the invitation token which defaults to 14 days in the future from creation").
			Default(func() time.Time {
				return time.Now().AddDate(0, 0, defaultInviteExpiresDays)
			}).
			Annotations(
				entgql.OrderField("expires"),
			).
			Optional(),
		field.String("recipient").
			Comment("the email used as input to generate the invitation token and is the destination person the invitation is sent to who is required to accept to join the organization").
			Validate(func(email string) error {
				_, err := mail.ParseAddress(email)
				return err
			}).
			Annotations(
				entx.FieldSearchable(),
			).
			Immutable().
			NotEmpty(),
		field.Enum("status").
			Comment("the status of the invitation").
			GoType(enums.InviteStatus("")).
			Annotations(
				entgql.OrderField("STATUS"),
			).
			Default(string(enums.InvitationSent)),
		field.Enum("role").
			GoType(enums.Role("")).
			Values(string(enums.RoleOwner)).
			Default(string(enums.RoleMember)),
		field.Int("send_attempts").
			Comment("the number of attempts made to perform email send of the invitation, maximum of 5").
			Annotations(
				entgql.OrderField("send_attempts"),
			).
			Default(1),
		field.String("requestor_id").
			Comment("the user who initiated the invitation").
			Immutable().
			Optional().
			NotEmpty(),
		field.Bytes("secret").
			Comment("the comparison secret to verify the token's signature").
			NotEmpty().
			Nillable().
			Annotations(entgql.Skip()).
			Sensitive(),
		field.Bool("ownership_transfer").
			Comment("indicates if this invitation is for transferring organization ownership - when accepted, current owner becomes admin and invitee becomes owner").
			Default(false).
			Optional(),
	}
}

// Mixin of the Invite
func (i Invite) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags: true,
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(i, withSkipTokenTypesObjects(&token.OrgInviteToken{})),
		},
	}.getMixins(i)
}

// Indexes of the Invite
func (Invite) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("recipient", ownerFieldName).
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Edges of the Invite
func (i Invite) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeToWithPagination(i, Event{}),
		defaultEdgeToWithPagination(i, Group{}),
	}
}

func (Invite) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogBaseModule,
	}
}

// Annotations of the Invite
func (i Invite) Annotations() []schema.Annotation {
	return []schema.Annotation{
		history.Annotations{
			Exclude: true,
		},
	}
}

// Hooks of the Invite
func (Invite) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookEmailValidation(),
		hooks.HookInvite(),
		hooks.HookInviteGroups(),
		hooks.HookInviteAccepted(),
	}
}

// Policy of the Invite
func (i Invite) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			rule.AllowIfContextHasPrivacyTokenOfType[*token.OrgInviteToken](),
		),
		policy.WithMutationRules(
			rule.AllowIfContextHasPrivacyTokenOfType[*token.OrgInviteToken](),
			rule.CanInviteUsers(),
			policy.CheckOrgWriteAccess(),
			rule.AllowMutationAfterApplyingOwnerFilter(),
		),
	)
}
