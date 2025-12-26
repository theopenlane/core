package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/common/enums"
	"github.com/theopenlane/entx/accessmap"
)

// UserSetting holds the schema definition for the User entity.
type UserSetting struct {
	SchemaFuncs

	ent.Schema
}

// SchemaUserSetting is the name of the UserSetting schema.
const SchemaUserSetting = "user_setting"

// Name returns the name of the UserSetting schema.
func (UserSetting) Name() string {
	return SchemaUserSetting
}

// GetType returns the type of the UserSetting schema.
func (UserSetting) GetType() any {
	return UserSetting.Type
}

// PluralName returns the plural name of the UserSetting schema.
func (UserSetting) PluralName() string {
	return pluralize.NewClient().Plural(SchemaUserSetting)
}

// Mixin of the UserSetting
func (UserSetting) Mixin() []ent.Mixin {
	return getDefaultMixins(UserSetting{})
}

// Fields of the UserSetting
func (UserSetting) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id").Optional(),
		field.Bool("locked").
			Comment("user account is locked if unconfirmed or explicitly locked").
			Default(false),
		field.Time("silenced_at").
			Comment("The time notifications regarding the user were silenced").
			Optional().
			Nillable(),
		field.Time("suspended_at").
			Comment("The time the user was suspended").
			Optional().
			Nillable(),
		field.Enum("status").
			Comment("status of the user account").
			GoType(enums.UserStatus("")).
			Default(string(enums.UserStatusActive)),
		field.Bool("email_confirmed").Default(false).
			Comment("whether the user has confirmed their email address"),
		field.Bool("is_webauthn_allowed").
			Comment("specifies a user may complete authentication by verifying a WebAuthn capable device").
			Annotations(
				entgql.Skip(^entgql.SkipType),
			).
			Optional().
			Default(false),
		field.Bool("is_tfa_enabled").
			Comment("whether the user has two factor authentication enabled").
			Optional().
			Default(false),
		field.String("phone_number").
			Comment("phone number associated with the account, used 2factor SMS authentication").
			Optional().
			Annotations(
				// skip until SMS 2fa feature is implemented
				entgql.Skip(entgql.SkipAll),
			).
			Nillable(),
	}
}

// Edges of the UserSetting
func (u UserSetting) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: u,
			edgeSchema: User{},
			ref:        "setting",
			field:      "user_id",
			annotations: []schema.Annotation{
				accessmap.EdgeNoAuthCheck(),
			},
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: u,
			name:       "default_org",
			t:          Organization.Type,
			comment:    "organization to load on user login",
			annotations: []schema.Annotation{
				accessmap.EdgeNoAuthCheck(),
			},
		}),
	}
}

// Hooks of the UserSetting.
func (UserSetting) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookUserSetting(),
		hooks.HookUserSettingEmailConfirmation(),
	}
}

// Interceptors of the UserSetting.
func (UserSetting) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorUserSetting(),
	}
}

func (UserSetting) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.AllowCreate(),
		),
		policy.WithOnMutationRules(
			ent.OpUpdateOne|ent.OpUpdate,
			rule.AllowIfContextHasPrivacyTokenOfType[*token.VerifyToken](),
			rule.AllowIfContextHasPrivacyTokenOfType[*token.OauthTooToken](),
			rule.AllowIfSelf(),
		),
	)
}
