package schema

import (
	"net/mail"
	"time"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	emixin "github.com/theopenlane/entx/mixin"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/internal/ent/validator"
	"github.com/theopenlane/core/pkg/enums"
)

const (
	urlMaxLen  = 2048
	nameMaxLen = 64
)

// User holds the schema definition for the User entity.
type User struct {
	SchemaFuncs

	ent.Schema
}

const SchemaUser = "user"

func (User) Name() string {
	return SchemaUser
}

func (User) GetType() any {
	return User.Type
}

func (User) PluralName() string {
	return pluralize.NewClient().Plural(SchemaUser)
}

// Mixin of the User
func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		mixin.SoftDeleteMixin{},
		emixin.IDMixin{
			HumanIdentifierPrefix: "USR",
			SingleFieldIndex:      true,
		},
		emixin.TagMixin{},
		mixin.GraphQLAnnotationMixin{},
	}
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		// NOTE: the created_at and updated_at fields are automatically created by the AuditMixin, you do not need to re-declare / add them in these fields
		field.String("email").
			Validate(func(email string) error {
				_, err := mail.ParseAddress(email)
				return err
			}),
		field.String("first_name").
			Optional().
			MaxLen(nameMaxLen).
			Annotations(
				entgql.OrderField("first_name"),
			),
		field.String("last_name").
			Optional().
			MaxLen(nameMaxLen).
			Annotations(
				entgql.OrderField("last_name"),
			),
		field.String("display_name").
			Comment("The user's displayed 'friendly' name").
			MaxLen(nameMaxLen).
			NotEmpty().
			Annotations(
				entgql.OrderField("display_name"),
			),
		field.String("avatar_remote_url").
			Comment("URL of the user's remote avatar").
			MaxLen(urlMaxLen).
			Validate(validator.ValidateURL()).
			Optional().
			Nillable(),
		field.String("avatar_local_file_id").
			Comment("The user's local avatar file id, takes precedence over the avatar remote URL").
			Optional().
			Annotations(
				// this field is not exposed to the graphql schema, it is set by the file upload handler
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			).
			Nillable(),
		field.Time("avatar_updated_at").
			Comment("The time the user's (local) avatar was last updated").
			Default(time.Now).
			UpdateDefault(time.Now).
			Optional().
			Nillable(),
		field.Time("last_seen").
			Comment("the time the user was last seen").
			UpdateDefault(time.Now).
			Optional().
			Nillable(),
		field.String("password").
			Comment("user password hash").
			Nillable().
			Sensitive().
			Optional(),
		field.String("sub").
			Comment("the Subject of the user JWT").
			Unique().
			Optional(),
		field.Enum("auth_provider").
			Comment("auth provider used to register the account").
			GoType(enums.AuthProvider("")).
			Default(string(enums.AuthProviderCredentials)),
		field.Enum("role").
			Comment("the user's role").
			GoType(enums.Role("")).
			Values(enums.RoleUser.String()). // add user as a role
			Default(enums.RoleUser.String()).
			Optional(),
	}
}

// Indexes of the User
func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email").
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Edges of the User
func (u User) Edges() []ent.Edge {
	return []ent.Edge{
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         u,
			edgeSchema:         PersonalAccessToken{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         u,
			edgeSchema:         TFASetting{},
			cascadeDeleteOwner: true,
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema:    u,
			name:          "setting",
			t:             UserSetting.Type,
			required:      true,
			cascadeDelete: "User",
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         u,
			edgeSchema:         EmailVerificationToken{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         u,
			edgeSchema:         PasswordResetToken{},
			cascadeDeleteOwner: true,
		}),

		edge.To("groups", Group.Type).
			Annotations(entgql.RelayConnection()).
			Through("group_memberships", GroupMembership.Type),
		edge.To("organizations", Organization.Type).
			Annotations(entgql.RelayConnection()).
			Through("org_memberships", OrgMembership.Type),

		edgeToWithPagination(&edgeDefinition{
			fromSchema:         u,
			name:               Webauthn{}.Name(),
			t:                  Webauthn.Type,
			cascadeDeleteOwner: true,
		}),

		defaultEdgeToWithPagination(u, File{}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: u,
			name:       "avatar_file",
			t:          File.Type,
			field:      "avatar_local_file_id",
		}),

		defaultEdgeToWithPagination(u, Event{}),
		defaultEdgeToWithPagination(u, ActionPlan{}),
		defaultEdgeToWithPagination(u, Subcontrol{}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: u,
			name:       "assigner_tasks",
			t:          Task.Type,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: u,
			name:       "assignee_tasks",
			t:          Task.Type,
		}),

		edge.To("programs", Program.Type).
			Through("program_memberships", ProgramMembership.Type),
	}
}

// Annotations of the User
func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// Delete users from groups and orgs when the user is deleted
		entx.CascadeThroughAnnotationField(
			[]entx.ThroughCleanup{
				{
					Field:   "User",
					Through: "OrgMembership",
				},
				{
					Field:   "User",
					Through: "GroupMembership",
				},
			},
		),
	}
}

// Policy of the User
func (User) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			// interceptors are setup to filter users outside of the organization
			privacy.AlwaysAllowRule(),
		),
		policy.WithOnMutationRules(
			// the user hook has update operations on user create so we need to allow email
			// token sign up for update operations as well
			ent.OpCreate|ent.OpUpdateOne,
			rule.AllowIfContextHasPrivacyTokenOfType[*token.SignUpToken](),
			rule.AllowIfContextHasPrivacyTokenOfType[*token.OrgInviteToken](),
			rule.AllowIfContextHasPrivacyTokenOfType[*token.OauthTooToken](),
			rule.AllowIfSelf(),
		),
		policy.WithOnMutationRules(
			ent.OpUpdate|ent.OpDeleteOne|ent.OpDelete,
			rule.AllowIfSelf(),
		),
	)
}

func (User) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookUser(),
		hooks.HookUserPermissions(),
		hooks.HookDeleteUser(),
	}
}

// Interceptors of the User.
func (d User) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.TraverseUser(),
	}
}
